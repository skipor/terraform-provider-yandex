package yandex

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/resourcemanager/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

const providerDefaultValueInsecure = false
const providerDefaultValuePlaintext = false
const providerDefaultValueEndpoint = "api.cloud.yandex.net:443"

var testAccProviders map[string]terraform.ResourceProvider

// WARNING!!!! do not use testAccProviderEmptyFolder in tests, that use testAccCheck***Destroy functions.
// testAccCheck***Destroy functions tend to use static testAccProvider
var testAccProviderEmptyFolder map[string]terraform.ResourceProvider

var testAccProvider *schema.Provider

var testAccEnvVars = []string{
	"YC_FOLDER_ID",
	"YC_CLOUD_ID",
	"YC_ZONE",
	"YC_TOKEN",
	"YC_LOGIN",
	"YC_LOGIN_2",
	"YC_STORAGE_ENDPOINT_URL",
	"YC_MESSAGE_QUEUE_ENDPOINT",
}

var testCloudID = "not initialized"
var testOrganizationID = "not initialized"
var testCloudName = "not initialized"
var testFolderID = "not initialized"
var testFolderName = "not initialized"
var testRoleID = "resource-manager.clouds.member"
var testUserLogin1 = "no user login"
var testUserLogin2 = "no user login"
var testUserID1 = "no user id"
var testUserID2 = "no user id"
var testStorageEndpoint = "no.storage.endpoint"

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"yandex": testAccProvider,
	}

	testAccProviderEmptyFolder = map[string]terraform.ResourceProvider{
		"yandex": emptyFolderProvider().(*schema.Provider),
	}

	if os.Getenv("TF_ACC") != "" {
		if err := setTestIDs(); err != nil {
			panic(err)
		}
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProviderWithRawConfig(t *testing.T) {
	testProvider := Provider().(*schema.Provider)

	raw := map[string]interface{}{
		"insecure": true,
		"token":    "any_string_like_a_oauth",
		"endpoint": "localhost:4433",
		"zone":     "ru-central1-a",
	}

	err := testProvider.Configure(terraform.NewResourceConfigRaw(raw))
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if err := testProvider.InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProviderDefaultValues(t *testing.T) {
	// save OS env vars
	envVars := []string{"YC_INSECURE", "YC_PLAINTEXT", "YC_ENDPOINT"}
	saveEnvVariable := saveAndUnsetEnvVars(envVars)

	testProvider := Provider().(*schema.Provider)

	raw := map[string]interface{}{
		"token": "any_string_like_a_oauth",
		"zone":  "ru-central1-a",
	}

	err := testProvider.Configure(terraform.NewResourceConfigRaw(raw))
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if err := testProvider.InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}

	conf := testProvider.Meta().(*Config)
	if conf.Endpoint != providerDefaultValueEndpoint {
		t.Errorf("there is not default API endpoint (%s), should be %s", conf.Endpoint, providerDefaultValueEndpoint)
	}

	if conf.Plaintext {
		t.Errorf("there is not default option 'Plaintext' (%v), should be %v", conf.Plaintext, providerDefaultValuePlaintext)
	}

	if conf.Insecure {
		t.Errorf("there is not default option 'Insecure' (%v), should be %v", conf.Plaintext, providerDefaultValueInsecure)
	}

	// restore OS env vars
	if err := restoreEnvVars(saveEnvVariable); err != nil {
		t.Fatal("failed to restore OS env vars:", envVars, "after test", t.Name(), " - error:", err)
	}
}

func testAccPreCheck(t *testing.T) {
	for _, varName := range testAccEnvVars {
		if val := os.Getenv(varName); val == "" {
			t.Fatalf("%s must be set for acceptance tests", varName)
		}
	}
}

func getExampleFolderName() string {
	return testFolderName
}

func getExampleCloudName() string {
	return testCloudName
}

func getExampleRoleID() string {
	return testRoleID
}

func getExampleCloudID() string {
	return testCloudID
}

func getExampleOrganizationID() string {
	return testOrganizationID
}

func getExampleFolderID() string {
	return testFolderID
}

func getExampleUserID1() string {
	return testUserID1
}

func getExampleUserID2() string {
	return testUserID2
}

func getExampleUserLogin1() string {
	return testUserLogin1
}

func getExampleUserLogin2() string {
	return testUserLogin2
}

func getExampleStorageEndpoint() string {
	return testStorageEndpoint
}

func setTestIDs() error {
	// init sdk client based on env var
	envEndpoint := os.Getenv("YC_ENDPOINT")
	if envEndpoint == "" {
		envEndpoint = defaultEndpoint
	}

	providerConfig := &Config{
		Token: os.Getenv("YC_TOKEN"),
	}
	credentials, err := providerConfig.credentials()
	if err != nil {
		return err
	}

	config := &ycsdk.Config{
		Credentials: credentials,
		Endpoint:    envEndpoint,
	}

	ctx := context.Background()

	sdk, err := ycsdk.Build(ctx, *config)
	if err != nil {
		return err
	}

	// setup example ID values for test cases
	testCloudID = os.Getenv("YC_CLOUD_ID")
	testOrganizationID = os.Getenv("YC_ORGANIZATION_ID")
	testCloudName = getCloudNameByID(sdk, testCloudID)

	testFolderID = os.Getenv("YC_FOLDER_ID")
	testFolderName = getFolderNameByID(sdk, testFolderID)

	testUserLogin1 = os.Getenv("YC_LOGIN")
	testUserLogin2 = os.Getenv("YC_LOGIN_2")

	testUserID1 = loginToUserID(sdk, testUserLogin1)
	testUserID2 = loginToUserID(sdk, testUserLogin2)

	testStorageEndpoint = os.Getenv("YC_STORAGE_ENDPOINT_URL")

	return nil
}

func getCloudNameByID(sdk *ycsdk.SDK, cloudID string) string {
	cloud, err := sdk.ResourceManager().Cloud().Get(context.Background(), &resourcemanager.GetCloudRequest{
		CloudId: cloudID,
	})
	if err != nil {
		log.Printf("could not get cloud name for %s: %s", cloudID, err)
		if reqID, ok := isRequestIDPresent(err); ok {
			log.Printf("[DEBUG] request ID is %s\n", reqID)
		}
		return "no cloud name detected"
	}
	return cloud.Name
}

func getFolderNameByID(sdk *ycsdk.SDK, folderID string) string {
	folder, err := sdk.ResourceManager().Folder().Get(context.Background(), &resourcemanager.GetFolderRequest{
		FolderId: folderID,
	})
	if err != nil {
		log.Printf("could not get folder name for %s: %s", folderID, err)
		if reqID, ok := isRequestIDPresent(err); ok {
			log.Printf("[DEBUG] request ID is %s\n", reqID)
		}
		return "no folder name detected"
	}
	return folder.Name
}

func loginToUserID(sdk *ycsdk.SDK, loginName string) (userID string) {
	account, err := sdk.IAM().YandexPassportUserAccount().GetByLogin(context.Background(), &iam.GetUserAccountByLoginRequest{
		Login: loginName,
	})
	if err != nil {
		log.Printf("could not get user Id for %s: %s", loginName, err)
		if reqID, ok := isRequestIDPresent(err); ok {
			log.Printf("[DEBUG] request ID is %s\n", reqID)
		}
		return "not initialized"
	}
	return account.Id
}

func saveAndUnsetEnvVars(varNames []string) map[string]string {
	storage := make(map[string]string, len(varNames))

	for _, v := range varNames {
		storage[v] = os.Getenv(v)
		os.Unsetenv(v)
	}

	return storage
}

func restoreEnvVars(storage map[string]string) error {
	for varName, varValue := range storage {
		if err := os.Setenv(varName, varValue); err != nil {
			return err
		}
	}
	return nil
}
