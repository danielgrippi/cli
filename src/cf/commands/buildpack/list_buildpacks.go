package buildpack

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
	"strconv"
)

type ListBuildpacks struct {
	ui            terminal.UI
	buildpackRepo api.BuildpackRepository
}

func NewListBuildpacks(ui terminal.UI, buildpackRepo api.BuildpackRepository) (cmd ListBuildpacks) {
	cmd.ui = ui
	cmd.buildpackRepo = buildpackRepo
	return
}

func (cmd ListBuildpacks) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
	}
	return
}

func (cmd ListBuildpacks) Run(c *cli.Context) {
	cmd.ui.Say("Getting buildpacks...\n")

	stopChan := make(chan bool)
	defer close(stopChan)

	buildpackChan, statusChan := cmd.buildpackRepo.ListBuildpacks(stopChan)

	table := cmd.ui.Table([]string{"buildpack", "position"})
	noBuildpacks := true

	for buildpacks := range buildpackChan {
		rows := [][]string{}
		for _, buildpack := range buildpacks {
			position := ""
			if buildpack.Position != nil {
				position = strconv.Itoa(*buildpack.Position)
			}
			rows = append(rows, []string{
				buildpack.Name,
				position,
			})
		}
		table.Print(rows)
		noBuildpacks = false
	}

	apiStatus := <-statusChan
	if apiStatus.IsNotSuccessful() {
		cmd.ui.Failed("Failed fetching buildpacks.\n%s", apiStatus.Message)
		return
	}

	if noBuildpacks {
		cmd.ui.Say("No buildpacks found")
	}
}
