package deployment

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/salmanrf/capybara-cloud/internal/application"
	"github.com/salmanrf/capybara-cloud/internal/database"
	config "github.com/salmanrf/capybara-cloud/pkg/utils"
	shared_deployment "github.com/salmanrf/capybara-cloud/shared/deployment"
)

func TestDeployExtract(t *testing.T) {
	ctx := context.Background()
	deployment_repository := &StubAppDeploymentRepository{}

	port_service := &shared_deployment.StubPortAllocatorService{}
	app_service := &application.StubApplicationService{}
	deployment_service := NewService(
		ctx,
		&StubDocker{},
		app_service,
		&StubMasbroService{},
		port_service,
		deployment_repository,
		make(chan DeployRequest, 1),
	)

	t.Run("should perform artifact extraction when status is DEPLOY_STATUS_INITIATED = 1", func (t *testing.T) {
		cfg := config.GetConfig()
		cfg.BASE_BUILD_PATH = "/tmp/tests/build-artifacts"
		config.SetConfig(cfg)
		
		// * Uses /deployment/test-samples/<filename>
		type test struct{
			tid   							uuid.UUID
			name 								string
			artifact_filename 	string
			artifact_path 			string
		}
		tests := []test{
			{
				uuid.New(),
				"abcd",
				"abcd.tar.gz",
				"",
			},
			{
				uuid.New(),
				"app-html",
				"sample-html.tar.gz",
				"",
			},
			{
				uuid.New(),
				"app-vue",
				"sample-vue.tar.gz",
				"",
			},
		}

		pwd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}

		// * For inspecting the content of a tarball
		// * Use output from tar to compare with ReadDir 
		/* Sample output:
			./
			./d.txt
			./b.txt
			./a.txt
			./c.txt
		*/
		tar, err := exec.LookPath("tar")
		if err != nil {
			t.Fatalf("got error %v, want nil", err)
		}


		// * Cleanup build directories
		defer func () {
			os.Remove(cfg.BASE_BUILD_PATH)
		}()

		for _, tt := range tests {
			t.Run("should match file count by tar -tzf after extraction", func (t *testing.T) {
				tt.artifact_path = path.Join(pwd, "test-samples", tt.artifact_filename)
				
				mock_app := database.Application{
					Name: tt.name,
					Type: shared_deployment.APP_TYPE_NODEJS_CONTAINER,
				}	

				want_build_path := getBuildDirPath(cfg, "localfs", mock_app.Name)
				want_file_count := 0

				// * Store output of tar -tzf
				output := bytes.NewBuffer([]byte{})
				cmd := exec.Command(tar, "-tzf", tt.artifact_path)
				cmd.Stdout = output
				if err := cmd.Run(); err != nil {
					t.Fatalf("got error checking actual tar content %v, want nil", err)
				}
				// * Clean trailing newline
				output_str := strings.Trim(output.String(), " \n")
				
				// * Format tar output, clean '.' and take only the first entry after the first slash
				// * eg, ./abcd.txt -> abcd.txt
				// * eg, ./a/b -> a
				// * Track occurence in map
				entries := strings.Split(output_str, "\n")
				file_entries := make([]string, 0)
				base_map := make(map[string]bool)
				for _, ent := range entries {
					dot_cleaned := strings.Trim(ent, ".")
					if dot_cleaned == "" {
						continue
					}
					
					parts :=  strings.Split(dot_cleaned, "/")
					if len(parts) < 2 {
						continue
					}

					clean := parts[1]
					if clean == "" {
						continue
					}
					if base_map[clean] {
						continue
					} else {
						base_map[clean] = true	
					}

					file_entries = append(file_entries, clean)
				}
				want_file_count = len(file_entries)

				mock_dp := database.ApplicationDeployment{
					AppID: mock_app.AppID,
					Status: shared_deployment.DEPLOY_STATUS_INITIATED,
					ArtifactsPath: tt.artifact_path,
					BuildPath: want_build_path,
					VersionNumber: 1,
				}
				mock_app_cfg := database.ApplicationConfig{
					Port: 3000,
				}

				deploy_request := DeployRequest{
					ApplicationDto: mock_app,
					DeploymentDto: mock_dp,
					ApplicationConfig: mock_app_cfg,
				}

				res, err := deployment_service.Extract(deploy_request)
				if err != nil {
					t.Errorf("got deployment build step error %v, want nil", err)
				}

				dep := res.DeploymentDto
				got_new_status := dep.Status
				want_new_status := shared_deployment.DEPLOY_STATUS_BUILD_EXTRACTED

				if got_new_status != int32(want_new_status) {
					t.Errorf("got new deployment status %d, want %d", got_new_status, want_new_status)
				}

				got_entries, err := os.ReadDir(want_build_path)
				if err != nil {
					t.Errorf("got read build dir %s error %v, want nil", want_build_path, err)
				}

				got_file_count := len(got_entries)
				if got_file_count != want_file_count {
					t.Errorf("got %d file count after extract, want %d", got_file_count, want_file_count)
				}
			})
		}
	})
}