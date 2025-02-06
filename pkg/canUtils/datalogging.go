package canUtils

import (
	"fmt"
	"os/exec"
	"sync"
)

var (
  cmd *exec.Cmd
)

func DoDatalogging(dataloggingRunning *bool, wg *sync.WaitGroup) {
	defer wg.Done()

  if *dataloggingRunning {
    fmt.Println("--- Datalogging: Start ---")
    cmd = exec.Command("/home/pi/dev/wt-datalogging/bin/wt-datalogging")

    output, err := cmd.Output()
    if err != nil {
      fmt.Println("Error running datalogging: ", err)
    }
    // TODO: This won't log within here until we exit, fix
    fmt.Println(string(output))

  } else {
    fmt.Println("--- Datalogging: Finish ---")

    err := cmd.Process.Kill()
    if err != nil {
      fmt.Println("Error closing datalogging: ", err)
    }
    return
  }
}