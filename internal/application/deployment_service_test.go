package application

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"os/exec"
	"path"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/moby/moby/client"
	"github.com/salmanrf/capybara-cloud/internal/database"
	"github.com/salmanrf/capybara-cloud/pkg/utils"
	"github.com/salmanrf/capybara-cloud/tests"
)

var mock_user_id = "df3e9f69-bbe4-4cbb-8479-dba87dd17e2c"
var mock_app_id = "d8dab736-e7a5-49f6-97a0-64670faa2a5d"
var mock_bundle = multipart.File(&nopCloser{})

func TestDeploy(t *testing.T) {
	ctx := context.Background()
	deployment_repository := &StubAppDeploymentRepository{}

	app_service := &tests.StubApplicationService{}
	deployment_service := NewDeploymentService(
		ctx,
		app_service,
		deployment_repository,
	)

	t.Run("should return error if application is not found", func (t *testing.T) {
		defer func () {
			app_service.Clear()
		}()

		dp, err := deployment_service.Deploy(mock_user_id, mock_app_id, nil, mock_bundle);

		got_dp := dp
		var want_dp *database.ApplicationDeployment = nil
		want_error := errors.New("not_found")
		got_error := err

		if got_dp != want_dp {
			t.Errorf("got app deployment %v, want %v", got_dp, want_dp)
		}
		
		if got_error.Error() != want_error.Error() {
			t.Errorf("got error %v, want %v", got_error, want_error)
		}
	})

	t.Run("should return error if the logged in user is not a project member", func (t *testing.T) {
		defer func () {
			app_service.Clear()
		}()

		app_service.Find_one_error = errors.New("permission_denied")

		dp, err := deployment_service.Deploy(mock_user_id, mock_app_id, nil, mock_bundle);

		got_dp := dp
		var want_dp *database.ApplicationDeployment = nil
		want_error := errors.New("permission_denied")
		got_error := err

		if got_dp != want_dp {
			t.Errorf("got app deployment %v, want %v", got_dp, want_dp)
		}
		
		if got_error.Error() != want_error.Error() {
			t.Errorf("got error %v, want %v", got_error, want_error)
		}
	})

	t.Run("should return error if application config is missing", func (t *testing.T) {
		defer func () {
			app_service.Clear()
		}()

		mock_app := &database.FindOneApplicationWithProjectMemberRow{}
		app_service.Find_one_return = mock_app
		mock_app.ApplicationConfig.AppCfgID.Scan("abcd")

		dp, err := deployment_service.Deploy(mock_user_id, mock_app_id, nil, mock_bundle);

		got_dp := dp
		var want_dp *database.ApplicationDeployment = nil
		want_error := errors.New("not_found")
		got_error := err

		if got_dp != want_dp {
			t.Errorf("got app deployment %v, want %v", got_dp, want_dp)
		}

		if got_error.Error() != want_error.Error() {
			t.Errorf("got error %v, want %v", got_error, want_error)
		}
	})

	t.Run("should create and return the application deployment db item", func (t *testing.T) {
		defer func () {
			app_service.Clear()
		}()
		
		mock_app := &database.FindOneApplicationWithProjectMemberRow{}
		mock_app.AppID = pgtype.UUID{}
		mock_app.AppID.Scan(mock_app_id)
		mock_app.Name = "Handsome Capybara"
		mock_app.Type = "web_app_container:nodejs"
		mock_app.ApplicationConfig = database.ApplicationConfig{}
		mock_app.ApplicationConfig.AppCfgID = pgtype.UUID{}
		mock_app.ApplicationConfig.AppCfgID.Scan(uuid.New().String())
		mock_app.ApplicationConfig.VariablesJson = []byte(`PORT=8080`)

		app_service.Find_one_return = mock_app

		expected_deployment := &database.ApplicationDeployment{
			ContainerName: fmt.Sprintf("abcd:%s:%s", mock_app_id, mock_app.Name),
		}
		expected_deployment.AppDpID.Scan(uuid.New())
		expected_deployment.AppID.Scan(mock_app_id)

		deployment_repository.create_return = expected_deployment

		dp, err := deployment_service.Deploy(mock_user_id, mock_app_id, &multipart.FileHeader{}, mock_bundle);

		got_dp := dp
		got_err := err

		if got_err != nil {
			t.Errorf("got error %v, want nil", got_err)
		}

		if dp == nil {
			t.Error("got deployment nil, want pointer")
		}

		if !reflect.DeepEqual(got_dp, expected_deployment) {
			t.Errorf("got deployment %v, want %v", got_dp, expected_deployment)
		}
	})

	t.Run("should extract the deployment bundle and start a docker container for type web_app_container:nodejs", func (t *testing.T) {
		defer func () {
			app_service.Clear()
		}()
		
		mock_app := &database.FindOneApplicationWithProjectMemberRow{}
		mock_app.AppID = pgtype.UUID{}
		mock_app.AppID.Scan(mock_app_id)
		mock_app.Name = "Handsome Capybara"
		mock_app.Type = "web_app_container:nodejs"
		mock_app.ApplicationConfig = database.ApplicationConfig{}
		mock_app.ApplicationConfig.AppCfgID = pgtype.UUID{}
		mock_app.ApplicationConfig.AppCfgID.Scan(uuid.New().String())
		mock_app.ApplicationConfig.VariablesJson = []byte(`PORT=8080`)

		app_service.Find_one_return = mock_app

		_, err := deployment_service.Deploy(mock_user_id, mock_app_id, nil, mock_bundle)
		got_deployment := deployment_repository.create_call_args[0]

		if err != nil {
			t.Errorf("got deploy error %v, want nil", err)
		}
		
		dockerpath, err := exec.LookPath("docker")
		if err != nil {
			t.Errorf("Unable to find 'docker' executable")
		}

		want_artifact_path := got_deployment.ArtifactsPath
		want_full_path := fmt.Sprintf("/tmp/capybara_cloud/%s", want_artifact_path)
		_, err = os.Stat(want_full_path)

		if err != nil {
			t.Errorf("got artifact dir check err %v, want nil", err)
		}

		want_container_name := got_deployment.ContainerName
		cmd := exec.Command(dockerpath, "ps", "--filter", fmt.Sprintf("name=%s", want_container_name))

		cmdstdout := bytes.NewBuffer([]byte{})
		cmd.Stdout = cmdstdout
		got_err := cmd.Run()
		if got_err != nil {
			t.Errorf("got docker ps error %v, want nil", got_err)
		}

		got_output := cmdstdout.String()
		if !strings.Contains(got_output, want_container_name) {
			t.Errorf("got docker ps output:\r\n %s, want includes:\r\n %s", got_output, want_container_name)
		}
	})
}

func TestGetFullArtifactPath(t *testing.T) {
	mock_app := database.FindOneApplicationWithProjectMemberRow{}
	mock_app.AppID = pgtype.UUID{}
	mock_app.AppID.Scan(mock_app_id)
	mock_app.Type = "web_app_container:nodejs"
	mock_app.ApplicationConfig = database.ApplicationConfig{}
	mock_app.ApplicationConfig.AppCfgID = pgtype.UUID{}
	mock_app.ApplicationConfig.AppCfgID.Scan(uuid.New().String())
	mock_app.ApplicationConfig.VariablesJson = []byte(`PORT=8080`)

	tests := []struct{
		appname string
		fileheaders *multipart.FileHeader
	}{
		{"Handsome Capybara", &multipart.FileHeader{Filename: "capybundle.tar.gz"}},
		{"123Tailung&Von Astrea456__greyR4t", &multipart.FileHeader{Filename: "bundle.tar.gz"}},
		{"Sophia Foundations FE", &multipart.FileHeader{Filename: "sophsoph.tar.gz"}},
		{"Selma Sharia Finance", &multipart.FileHeader{Filename: "selmafin.tar.gz"}},
	}

	basepath := root_localfs_artifacts + "/artifact"

	for _, tt := range tests {
		t.Run(fmt.Sprintf("should format app name '%s' into artifact path", tt.appname), func (t *testing.T) {
			mock_app.Name = tt.appname
			appnameslug := utils.Slugify(tt.appname)

			got_path := getArtifactDirPath("localfs", mock_app)

			pattern, _ := regexp.Compile(fmt.Sprintf("%s-%s-\\d{4}-\\d{2}-\\d{2}-\\d{2}-\\d{2}-\\d{2}", basepath, appnameslug))
			if !pattern.Match([]byte(got_path)) {
				t.Errorf("got match false (%s), want true", got_path)
			}
		})
	}
}

func TestGetContainerName(t *testing.T) {
	mock_app := database.FindOneApplicationWithProjectMemberRow{}
	mock_app.AppID = pgtype.UUID{}
	mock_app.AppID.Scan(mock_app_id)
	mock_app.Type = "web_app_container:nodejs"
	mock_app.ApplicationConfig = database.ApplicationConfig{}
	mock_app.ApplicationConfig.AppCfgID = pgtype.UUID{}
	mock_app.ApplicationConfig.AppCfgID.Scan(uuid.New().String())
	mock_app.ApplicationConfig.VariablesJson = []byte(`PORT=8080`)

	tests := []struct{
		appname string
	}{
		{"Handsome Capybara"},
		{"123Tailung&Von Astrea456__greyR4t"},
		{"Sophia Foundations FE"},
		{"Selma Sharia Finance"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("should format app name '%s' into artifact path", tt.appname), func (t *testing.T) {
			mock_app.Name = tt.appname
			appnameslug := utils.Slugify(tt.appname)

			got_path := getContainerName(mock_app)

			pattern, _ := regexp.Compile(fmt.Sprintf("%s-%s-\\d{4}-\\d{2}-\\d{2}-\\d{2}-\\d{2}-\\d{2}", mock_app.Type, appnameslug))
			if !pattern.Match([]byte(got_path)) {
				t.Errorf("got match false (%s), want true", got_path)
			}
		})
	}
}

func TestSaveDeployArtifacts(t *testing.T) {
	mock_app := &database.FindOneApplicationWithProjectMemberRow{}
	mock_app.AppID = pgtype.UUID{}
	mock_app.AppID.Scan(mock_app_id)
	mock_app.Type = "web_app_container:nodejs"
	mock_app.ApplicationConfig = database.ApplicationConfig{}
	mock_app.ApplicationConfig.AppCfgID = pgtype.UUID{}
	mock_app.ApplicationConfig.AppCfgID.Scan(uuid.New().String())
	mock_app.ApplicationConfig.VariablesJson = []byte(`PORT=8080`)

	tests := []struct{
		appname string
		samplefilename string
	}{
		{"Handsome Capybara", "node-express.tar.gz"},
		{"123Tailung&Von Astrea456__greyR4t", "node-express.tar.gz"},
		{"Sophia Foundations FE", "node-express.tar.gz"},
		{"Selma Sharia Finance", "node-express.tar.gz"},
	}

	defer func () {
		err := os.RemoveAll("/tmp/capybara_cloud")
		if err != nil {
			fmt.Println("Cleanup error:", err)
		}
	}()

	pwd, _ := os.Getwd()

	for _, tt := range tests {
		t.Run("should create the artifact directory and return the path given an app and bundle file", func (t *testing.T) {
			mock_app := *mock_app
			mock_app.Name = tt.appname
					
			samplefile, _ := os.OpenFile(path.Join(pwd, "samples", tt.samplefilename), os.O_RDONLY, 0) 
			fstat, _ := samplefile.Stat()
			fheaders := &multipart.FileHeader{
				Filename: fstat.Name(),
				Size: fstat.Size(),
			}

			defer func () {
				samplefile.Close()
			}()
			
			want_dir_path := getArtifactDirPath("", mock_app)			
			want_full_path := want_dir_path + "/" + tt.samplefilename
			got_path, _ := saveDeployArtifacts(mock_app, fheaders, samplefile)

			if got_path != want_full_path {
				t.Errorf("got artifact path %s, want %s", got_path, want_dir_path)
			}

			_, err := os.Stat(want_dir_path)
			if err != nil {
				t.Errorf("got artifact dir stat error %v, want nil", err)
			}

			_, err = os.Stat(want_full_path)
			if err != nil {
				t.Errorf("got bundle artifact stat error %v, want nil", err)
			}
		})
	}
}

func TestStartContainerizedApp(t *testing.T) {
	ctx := context.Background()
	
	globalConfig := utils.Config{
		DOCKER_REGISTRY: "capybara-cloud-tests",
	} 
	
	mock_app := database.FindOneApplicationWithProjectMemberRow{}
	mock_app.AppID = pgtype.UUID{}
	mock_app.AppID.Scan(mock_app_id)
	mock_app.Type = "container_nodejs"
	mock_app.ApplicationConfig = database.ApplicationConfig{}
	mock_app.ApplicationConfig.AppCfgID = pgtype.UUID{}
	mock_app.ApplicationConfig.AppCfgID.Scan(uuid.New().String())
	mock_app.ApplicationConfig.VariablesJson = []byte(`PORT=8080`)
	mock_deployment := database.ApplicationDeployment{}

	tests := []struct{
		appname string
		samplefilename string
	}{
		{"Handsome Capybara", "node-express.tar.gz"},
		{"123Tailung&Von Astrea456__greyR4t", "node-express.tar.gz"},
		{"Sophia Foundations FE", "node-express.tar.gz"},
		{"Selma Sharia Finance", "node-express.tar.gz"},
	}

	pwd, _ := os.Getwd()

	dockerClient, err := client.New(client.FromEnv)
	if err != nil {
		log.Fatal(err)
	}

	for _, tt := range tests {
		ttmock_app := mock_app
		ttmock_app.Name = tt.appname
		ttmock_dp := mock_deployment
		ttmock_dp.ContainerName = getContainerName(ttmock_app)
		ttmock_dp.ArtifactsPath = getArtifactDirPath("", ttmock_app)
		
		t.Run(fmt.Sprintf("should start the containerized app %s", getContainerName(ttmock_app)), func (t *testing.T) {
			samplefile, samplefileerr := os.OpenFile(path.Join(pwd, "samples", tt.samplefilename), os.O_RDONLY, 0) 
			if samplefileerr != nil {
				t.Errorf("got open sample file err %v, want nil", samplefileerr)
			}
			fstat, fstaterr := samplefile.Stat()
			if fstaterr != nil {
				t.Errorf("got sample file stat err %v, want nil", fstaterr)
			}
			fheaders := &multipart.FileHeader{
				Filename: fstat.Name(),
				Size: fstat.Size(),
			}
			ttmock_dp.ArtifactsPath += "/" + fheaders.Filename

			defer func () {
				samplefile.Close()
				_, err := dockerClient.ContainerRemove(
					ctx,
					ttmock_dp.ContainerName,
					client.ContainerRemoveOptions{
						Force: true,
					},
				)
				if err != nil {
					t.Log("Error cleaning up containers", err)
				}
				err = os.RemoveAll("/tmp/capybara_cloud")
				if err != nil {
					fmt.Println("Cleanup error:", err)
				}
			}()
			saveDeployArtifacts(ttmock_app, fheaders, samplefile)

			got_err := startContainerizedApp(globalConfig, ttmock_app, ttmock_dp, mock_app.ApplicationConfig)

			if got_err != nil {
				t.Errorf("got start container error %v, want nil", got_err)
			}
			
			want_container_name := ttmock_dp.ContainerName

			containerFilters := client.Filters{}
			containerFilters.Add("name", want_container_name)
			containerList, err := dockerClient.ContainerList(t.Context(), client.ContainerListOptions{
				Filters: containerFilters,
			})
			if err != nil {
				t.Errorf("got container list error %v, want nil", err)
			}
			
			got_found := false
			want_found := true
			for _, ct := range containerList.Items {
				for _, n := range ct.Names {
					if strings.Contains(n, want_container_name) {
						got_found = true
						break
					}
					if got_found {
						break
					}
				}
			}

			if got_found != want_found {
				t.Errorf("got container '%s' found == %v, want %v", want_container_name, got_found, want_found)
			}
		})
	}
}



type nopCloser struct {
	io.ReadSeekCloser
	io.ReaderAt
}

func (nopCloser) Close() error { return nil }