package configor

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/asaskevich/govalidator"
	"gopkg.in/yaml.v2"
)

type Anonymous struct {
	Description string
}

type Database struct {
	Name     string
	User     string `yaml:",omitempty" default:"root"`
	Password string `required:"true" env:"DBPassword"`
	Port     uint   `default:"3306" yaml:",omitempty" json:",omitempty"`
	SSL      bool   `default:"true" yaml:",omitempty" json:",omitempty"`
}

type Contact struct {
	Name  string `default:"sumo" yaml:",omitempty"  json:",omitempty"`
	Email string `required:"true"`
}

type testConfig struct {
	APPName   string `default:"configor" yaml:",omitempty" json:",omitempty"`
	Hosts     []string
	DB        *Database
	Contacts  []Contact
	Anonymous `anonymous:"true"`
	private   string
}

func generateDefaultConfig() testConfig {
	return testConfig{
		APPName: "configor",
		Hosts:   []string{"http://example.org", "http://jinzhu.me"},
		DB: &Database{
			Name:     "configor",
			User:     "configor",
			Password: "configor",
			Port:     3306,
			SSL:      true,
		},
		Contacts: []Contact{
			{
				Name:  "sumo",
				Email: "sumo@gmail.com",
			},
			{
				Name:  "sumo2",
				Email: "sumo2@gmail.com",
			},
		},
		Anonymous: Anonymous{
			Description: "This is an anonymous embedded struct whose environment variables should NOT include 'ANONYMOUS'",
		},
	}
}

func TestLoadNormaltestConfig(t *testing.T) {
	config := generateDefaultConfig()
	if bytes, err := json.Marshal(config); err == nil {
		if file, err := ioutil.TempFile("/tmp", "configor"); err == nil {
			defer file.Close()
			defer os.Remove(file.Name())
			file.Write(bytes)

			var result testConfig
			Load(&result, file.Name())
			if !reflect.DeepEqual(result, config) {
				t.Errorf("result should equal to original configuration")
			}
		}
	} else {
		t.Errorf("failed to marshal config")
	}
}

// CONFIGOR_DEBUG_MODE=true CONFIGOR_VERBOSE_MODE=true go test -v   -run TestDefaultValue -count=1
func TestDefaultValue(t *testing.T) {
	config := generateDefaultConfig()
	config.APPName = ""
	config.DB.Port = 0
	config.DB.SSL = false
	config.Contacts[0].Name = ""

	if bytes, err := json.Marshal(config); err == nil {
		if file, err := ioutil.TempFile("/tmp", "configor"); err == nil {
			defer file.Close()
			defer os.Remove(file.Name())
			file.Write(bytes)

			var result testConfig
			if err := Load(&result, file.Name()); err != nil {
				t.Error(err)
			}

			if !reflect.DeepEqual(result, generateDefaultConfig()) {
				t.Errorf("\nExpected: %+v, \nGot: %+v", generateDefaultConfig(), result)
			}
		}
	} else {
		t.Errorf("failed to marshal config")
	}
}

func TestMissingRequiredValue(t *testing.T) {
	config := generateDefaultConfig()
	config.DB.Password = ""

	if bytes, err := json.Marshal(config); err == nil {
		if file, err := ioutil.TempFile("/tmp", "configor"); err == nil {
			defer file.Close()
			defer os.Remove(file.Name())
			file.Write(bytes)

			var result testConfig
			if err := Load(&result, file.Name()); err == nil {
				t.Errorf("Should got error when load configuration missing db password")
			}
		}
	} else {
		t.Errorf("failed to marshal config")
	}
}

func TestUnmatchedKeyInYamltestConfigFile(t *testing.T) {
	type configStruct struct {
		Name string
	}
	type configFile struct {
		Name string
		Test string
	}
	config := configFile{Name: "test", Test: "ATest"}

	file, err := ioutil.TempFile("/tmp", "configor")
	if err != nil {
		t.Fatal("Could not create temp file")
	}

	defer os.Remove(file.Name())
	defer file.Close()

	filename := file.Name()

	if data, err := yaml.Marshal(config); err == nil {
		file.WriteString(string(data))

		var result configStruct

		// Do not return error when there are unmatched keys but ErrorOnUnmatchedKeys is false
		if err := New(&Config{}).Load(&result, filename); err != nil {
			t.Errorf("Should NOT get error when loading configuration with extra keys. Error: %v", err)
		}

		// Return an error when there are unmatched keys and ErrorOnUnmatchedKeys is true
		if err := New(&Config{ErrorOnUnmatchedKeys: true}).Load(&result, filename); err == nil {
			t.Errorf("Should get error when loading configuration with extra keys")

			// The error should be of type *yaml.TypeError
		} else if _, ok := err.(*yaml.TypeError); !ok {
			// || !strings.Contains(err.Error(), "not found in struct") {
			t.Errorf("Error should be of type yaml.TypeError. Instead error is %v", err)
		}

	} else {
		t.Errorf("failed to marshal config")
	}

	// Add .yaml to the file name and test again
	err = os.Rename(filename, filename+".yaml")
	if err != nil {
		t.Errorf("Could not add suffix to file")
	}
	filename = filename + ".yaml"
	defer os.Remove(filename)

	var result configStruct

	// Do not return error when there are unmatched keys but ErrorOnUnmatchedKeys is false
	if err := New(&Config{}).Load(&result, filename); err != nil {
		t.Errorf("Should NOT get error when loading configuration with extra keys. Error: %v", err)
	}

	// Return an error when there are unmatched keys and ErrorOnUnmatchedKeys is true
	if err := New(&Config{ErrorOnUnmatchedKeys: true}).Load(&result, filename); err == nil {
		t.Errorf("Should get error when loading configuration with extra keys")

		// The error should be of type *yaml.TypeError
	} else if _, ok := err.(*yaml.TypeError); !ok {
		// || !strings.Contains(err.Error(), "not found in struct") {
		t.Errorf("Error should be of type yaml.TypeError. Instead error is %v", err)
	}
}

func TestYamlDefaultValue(t *testing.T) {
	config := generateDefaultConfig()
	config.APPName = ""
	config.DB.Port = 0
	config.DB.SSL = false

	if bytes, err := yaml.Marshal(config); err == nil {
		if file, err := ioutil.TempFile("/tmp", "configor.*.yaml"); err == nil {
			defer file.Close()
			defer os.Remove(file.Name())
			file.Write(bytes)

			var result testConfig
			Load(&result, file.Name())

			if !reflect.DeepEqual(result, generateDefaultConfig()) {
				t.Errorf("result should be set default value correctly")
			}
		}
	} else {
		t.Errorf("failed to marshal config")
	}
}

func TestLoadtestConfigurationByEnvironment(t *testing.T) {
	config := generateDefaultConfig()
	config2 := struct {
		APPName string
	}{
		APPName: "config2",
	}

	if file, err := ioutil.TempFile("/tmp", "configor"); err == nil {
		defer file.Close()
		defer os.Remove(file.Name())
		configBytes, _ := yaml.Marshal(config)
		config2Bytes, _ := yaml.Marshal(config2)
		ioutil.WriteFile(file.Name()+".yaml", configBytes, 0644)
		defer os.Remove(file.Name() + ".yaml")
		ioutil.WriteFile(file.Name()+".production.yaml", config2Bytes, 0644)
		defer os.Remove(file.Name() + ".production.yaml")

		var result testConfig
		os.Setenv("CONFIGOR_ENV", "production")
		defer os.Setenv("CONFIGOR_ENV", "")
		if err := Load(&result, file.Name()+".yaml"); err != nil {
			t.Errorf("No error should happen when load configurations, but got %v", err)
		}

		var defaultConfig = generateDefaultConfig()
		defaultConfig.APPName = "config2"
		if !reflect.DeepEqual(result, defaultConfig) {
			t.Errorf("result should be load configurations by environment correctly")
		}
	}
}

func TestLoadtestConfigurationByEnvironmentSetBytestConfig(t *testing.T) {
	config := generateDefaultConfig()
	config2 := struct {
		APPName string
	}{
		APPName: "production_config2",
	}

	if file, err := ioutil.TempFile("/tmp", "configor"); err == nil {
		defer file.Close()
		defer os.Remove(file.Name())
		configBytes, _ := yaml.Marshal(config)
		config2Bytes, _ := yaml.Marshal(config2)
		ioutil.WriteFile(file.Name()+".yaml", configBytes, 0644)
		defer os.Remove(file.Name() + ".yaml")
		ioutil.WriteFile(file.Name()+".production.yaml", config2Bytes, 0644)
		defer os.Remove(file.Name() + ".production.yaml")

		var result testConfig
		var Configor = New(&Config{Environment: "production"})
		if Configor.Load(&result, file.Name()+".yaml"); err != nil {
			t.Errorf("No error should happen when load configurations, but got %v", err)
		}

		var defaultConfig = generateDefaultConfig()
		defaultConfig.APPName = "production_config2"
		if !reflect.DeepEqual(result, defaultConfig) {
			t.Errorf("result should be load configurations by environment correctly")
		}

		if Configor.GetEnvironment() != "production" {
			t.Errorf("configor's environment should be production")
		}
	}
}

func TestOverwritetestConfigurationWithEnvironmentWithDefaultPrefix(t *testing.T) {
	config := generateDefaultConfig()

	if bytes, err := json.Marshal(config); err == nil {
		if file, err := ioutil.TempFile("/tmp", "configor"); err == nil {
			defer file.Close()
			defer os.Remove(file.Name())
			file.Write(bytes)
			var result testConfig
			os.Setenv("CONFIGOR_APPNAME", "config2")
			os.Setenv("CONFIGOR_HOSTS", "- http://example.org\n- http://jinzhu.me")
			os.Setenv("CONFIGOR_DB_NAME", "db_name")
			defer os.Setenv("CONFIGOR_APPNAME", "")
			defer os.Setenv("CONFIGOR_HOSTS", "")
			defer os.Setenv("CONFIGOR_DB_NAME", "")
			Load(&result, file.Name())

			var defaultConfig = generateDefaultConfig()
			defaultConfig.APPName = "config2"
			defaultConfig.Hosts = []string{"http://example.org", "http://jinzhu.me"}
			defaultConfig.DB.Name = "db_name"
			if !reflect.DeepEqual(result, defaultConfig) {
				t.Errorf("result should equal to original configuration")
			}
		}
	}
}

func TestOverwritetestConfigurationWithEnvironment(t *testing.T) {
	config := generateDefaultConfig()

	if bytes, err := json.Marshal(config); err == nil {
		if file, err := ioutil.TempFile("/tmp", "configor"); err == nil {
			defer file.Close()
			defer os.Remove(file.Name())
			file.Write(bytes)
			var result testConfig
			os.Setenv("CONFIGOR_ENV_PREFIX", "app")
			os.Setenv("APP_APPNAME", "config2")
			os.Setenv("APP_DB_NAME", "db_name")
			defer os.Setenv("CONFIGOR_ENV_PREFIX", "")
			defer os.Setenv("APP_APPNAME", "")
			defer os.Setenv("APP_DB_NAME", "")
			Load(&result, file.Name())

			var defaultConfig = generateDefaultConfig()
			defaultConfig.APPName = "config2"
			defaultConfig.DB.Name = "db_name"
			if !reflect.DeepEqual(result, defaultConfig) {
				t.Errorf("result should equal to original configuration")
			}
		}
	}
}

func TestOverwritetestConfigurationWithEnvironmentThatSetBytestConfig(t *testing.T) {
	config := generateDefaultConfig()

	if bytes, err := json.Marshal(config); err == nil {
		if file, err := ioutil.TempFile("/tmp", "configor"); err == nil {
			defer file.Close()
			defer os.Remove(file.Name())
			file.Write(bytes)
			os.Setenv("APP1_APPName", "config2")
			os.Setenv("APP1_DB_Name", "db_name")
			defer os.Setenv("APP1_APPName", "")
			defer os.Setenv("APP1_DB_Name", "")

			var result testConfig
			var Configor = New(&Config{ENVPrefix: "APP1"})
			Configor.Load(&result, file.Name())

			var defaultConfig = generateDefaultConfig()
			defaultConfig.APPName = "config2"
			defaultConfig.DB.Name = "db_name"
			if !reflect.DeepEqual(result, defaultConfig) {
				t.Errorf("result should equal to original configuration")
			}
		}
	}
}

func TestResetPrefixToBlank(t *testing.T) {
	config := generateDefaultConfig()

	if bytes, err := json.Marshal(config); err == nil {
		if file, err := ioutil.TempFile("/tmp", "configor"); err == nil {
			defer file.Close()
			defer os.Remove(file.Name())
			file.Write(bytes)
			var result testConfig
			os.Setenv("CONFIGOR_ENV_PREFIX", "-")
			os.Setenv("APPNAME", "config2")
			os.Setenv("DB_NAME", "db_name")
			defer os.Setenv("CONFIGOR_ENV_PREFIX", "")
			defer os.Setenv("APPNAME", "")
			defer os.Setenv("DB_NAME", "")
			Load(&result, file.Name())

			var defaultConfig = generateDefaultConfig()
			defaultConfig.APPName = "config2"
			defaultConfig.DB.Name = "db_name"
			if !reflect.DeepEqual(result, defaultConfig) {
				t.Errorf("result should equal to original configuration")
			}
		}
	}
}

func TestResetPrefixToBlank2(t *testing.T) {
	config := generateDefaultConfig()

	if bytes, err := json.Marshal(config); err == nil {
		if file, err := ioutil.TempFile("/tmp", "configor"); err == nil {
			defer file.Close()
			defer os.Remove(file.Name())
			file.Write(bytes)
			var result testConfig
			os.Setenv("CONFIGOR_ENV_PREFIX", "-")
			os.Setenv("APPName", "config2")
			os.Setenv("DB_Name", "db_name")
			defer os.Setenv("CONFIGOR_ENV_PREFIX", "")
			defer os.Setenv("APPName", "")
			defer os.Setenv("DB_Name", "")
			Load(&result, file.Name())

			var defaultConfig = generateDefaultConfig()
			defaultConfig.APPName = "config2"
			defaultConfig.DB.Name = "db_name"
			if !reflect.DeepEqual(result, defaultConfig) {
				t.Errorf("result should equal to original configuration")
			}
		}
	}
}

func TestReadFromEnvironmentWithSpecifiedEnvName(t *testing.T) {
	config := generateDefaultConfig()

	if bytes, err := json.Marshal(config); err == nil {
		if file, err := ioutil.TempFile("/tmp", "configor"); err == nil {
			defer file.Close()
			defer os.Remove(file.Name())
			file.Write(bytes)
			var result testConfig
			os.Setenv("DBPassword", "db_password")
			defer os.Setenv("DBPassword", "")
			Load(&result, file.Name())

			var defaultConfig = generateDefaultConfig()
			defaultConfig.DB.Password = "db_password"
			if !reflect.DeepEqual(result, defaultConfig) {
				t.Errorf("result should equal to original configuration")
			}
		}
	}
}

func TestAnonymousStruct(t *testing.T) {
	config := generateDefaultConfig()

	if bytes, err := json.Marshal(config); err == nil {
		if file, err := ioutil.TempFile("/tmp", "configor"); err == nil {
			defer file.Close()
			defer os.Remove(file.Name())
			file.Write(bytes)
			var result testConfig
			os.Setenv("CONFIGOR_DESCRIPTION", "environment description")
			defer os.Setenv("CONFIGOR_DESCRIPTION", "")
			Load(&result, file.Name())

			var defaultConfig = generateDefaultConfig()
			defaultConfig.Anonymous.Description = "environment description"
			if !reflect.DeepEqual(result, defaultConfig) {
				t.Errorf("result should equal to original configuration")
			}
		}
	}
}

func TestENV(t *testing.T) {
	if ENV() != "test" {
		t.Errorf("Env should be test when running `go test`, instead env is %v", ENV())
	}

	os.Setenv("CONFIGOR_ENV", "production")
	defer os.Setenv("CONFIGOR_ENV", "")
	if ENV() != "production" {
		t.Errorf("Env should be production when set it with CONFIGOR_ENV")
	}
}

type slicetestConfig struct {
	Test1 int
	Test2 []struct {
		Test2Ele1 int
		Test2Ele2 int
	}
}

func TestSliceFromEnv(t *testing.T) {
	var tc = slicetestConfig{
		Test1: 1,
		Test2: []struct {
			Test2Ele1 int
			Test2Ele2 int
		}{
			{
				Test2Ele1: 1,
				Test2Ele2: 2,
			},
			{
				Test2Ele1: 3,
				Test2Ele2: 4,
			},
		},
	}

	var result slicetestConfig
	os.Setenv("CONFIGOR_TEST1", "1")
	os.Setenv("CONFIGOR_TEST2_0_TEST2ELE1", "1")
	os.Setenv("CONFIGOR_TEST2_0_TEST2ELE2", "2")

	os.Setenv("CONFIGOR_TEST2_1_TEST2ELE1", "3")
	os.Setenv("CONFIGOR_TEST2_1_TEST2ELE2", "4")
	err := Load(&result)
	if err != nil {
		t.Fatalf("load from env err:%v", err)
	}

	if !reflect.DeepEqual(result, tc) {
		t.Fatalf("unexpected result:%+v", result)
	}
}

func TestConfigFromEnv(t *testing.T) {
	type config struct {
		LineBreakString string `required:"true"`
		Count           int64
		Slient          bool
	}

	cfg := &config{}

	os.Setenv("CONFIGOR_ENV_PREFIX", "CONFIGOR")
	os.Setenv("CONFIGOR_LineBreakString", "Line one\nLine two\nLine three\nAnd more lines")
	os.Setenv("CONFIGOR_Slient", "1")
	os.Setenv("CONFIGOR_Count", "10")
	Load(cfg)

	if os.Getenv("CONFIGOR_LineBreakString") != cfg.LineBreakString {
		t.Error("Failed to load value has line break from env")
	}

	if !cfg.Slient {
		t.Error("Failed to load bool from env")
	}

	if cfg.Count != 10 {
		t.Error("Failed to load number from env")
	}
}

func TestValidation(t *testing.T) {
	type config struct {
		Name     string `valid:"-"`
		Title    string `valid:"alphanum,required"`
		AuthorIP string `valid:"ipv4"`
		Email    string `valid:"email"`
		Email2   string `valid:"email,optional"` // FIXME bug in govalidator
		Endpoint  string `valid:"required"`
		Count    int64
		Slient   bool
	}

	cfg := &config{Email: "a@b.com", Email2: "", AuthorIP: "1.1"}
	err := Load(cfg)
	fmt.Printf("%+v\n", cfg)
	if err != nil {
		errs := govalidator.ErrorsByField(err)
		cnt := 0
		for field, err := range errs {
			cnt++
			fmt.Printf("\t%d. %s: %s\n", cnt, field, err)
		}
		// t.Error("Error validating")
	}
}

func TestValidationMore(t *testing.T){

	type Address struct {
		Street string `valid:"-"`
		Zip    string `json:"zip" valid:"numeric,required"`
	}

	type User struct {
		Name     string `valid:"required"` // FIXME: https://github.com/asaskevich/govalidator/issues/359
		Email    string `valid:"required,email"`
		Password string `valid:"required"`
		Age      int    `valid:"required,numeric,range(1|200),@#\u0000"`
		Home     *Address
		Work     []Address
	}

	var tests = []struct {
		param    interface{}
		expected bool
	}{
		{&User{"John", "john@yahoo.com", "123G#678", 20, &Address{"Street", "ABC456D89"}, []Address{{"Street", "123456"}, {"Street", "123456"}}}, false},
		{&User{"John", "john!yahoo.com", "12345678", 20, &Address{"Street", "ABC456D89"}, []Address{{"Street", "ABC456D89"}, {"Street", "123456"}}}, false},
		{&User{"John", "", "12345", 0, &Address{"Street", "123456789"}, []Address{{"Street", "ABC456D89"}, {"Street", "123456"}}}, false},
		{&User{"", "john@yahoo.com", "123G#678", 20, &Address{"Street", "95504"}, []Address{{"Street", "123456"}, {"Street", "123456"}}}, true},
	}
	for _, test := range tests {
		err := Load(test.param)
		t.Log(err)
		if test.expected && err != nil {
				t.Errorf("Got Error on ValidateStruct(%#v)\n Error: %s", test.param, err)
		}
	}
}

func TestUsePkger(t *testing.T) {
	config := generateDefaultConfig()
	if bytes, err := json.Marshal(config); err == nil {
		if file, err := ioutil.TempFile(".", "temp_configor"); err == nil {
			defer file.Close()
			defer os.Remove(file.Name())
			file.Write(bytes)

			var result testConfig
			New(&Config{UsePkger: true}).Load(&result, "/"+file.Name())
			if !reflect.DeepEqual(result, config) {
				t.Errorf("result should equal to original configuration")
			}
		}
	} else {
		t.Errorf("failed to marshal config")
	}
}
