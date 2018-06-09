package setup

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/therecipe/qt/internal/binding/parser"
	"github.com/therecipe/qt/internal/binding/templater"

	"github.com/therecipe/qt/internal/cmd"
	"github.com/therecipe/qt/internal/utils"
)

func Generate(target string, docker, vagrant bool) {
	utils.Log.Infof("running: 'qtsetup generate %v' [docker=%v] [vagrant=%v]", target, docker, vagrant)
	if docker {
		cmd.Docker([]string{"/home/user/work/bin/qtsetup", "-debug", "generate"}, "linux", "", true)
		return
	}

	parser.LoadModules()

	for _, module := range parser.GetLibs() {
		if !parser.ShouldBuildForTarget(module, target) {
			utils.Log.Debugf("skipping generation of %v for %v", module, target)
			continue
		}

		mode := "full"
		switch {
		case target == "js":

		case target != runtime.GOOS:
			mode = "cgo"

		case utils.QT_STUB():
			mode = "stub"
		}
		utils.Log.Infof("generating %v qt/%v", mode, strings.ToLower(module))

		if target == runtime.GOOS || utils.QT_FAT() || (mode == "full" && target == "js") { //TODO: REVIEW
			templater.GenModule(module, target, templater.NONE)
		} else {
			templater.CgoTemplate(module, "", target, templater.MINIMAL, "", "") //TODO: collect errors
		}

		if utils.QT_DYNAMIC_SETUP() && mode == "full" && target != "js" {
			cc, _ := templater.ParseCgo(strings.ToLower(module), target)
			if cc != "" {
				cmd := exec.Command("go", "tool", "cgo", utils.GoQtPkgPath(strings.ToLower(module), strings.ToLower(module)+".go"))
				cmd.Dir = utils.GoQtPkgPath(strings.ToLower(module))
				utils.RunCmdOptional(cmd, fmt.Sprintf("failed to run cgo for %v (%v) on %v", target, strings.ToLower(module), runtime.GOOS))
			}
		}
	}
}
