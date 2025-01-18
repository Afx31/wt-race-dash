package datalogging

import (
	"fmt"
	"os/exec"
	"sync"
)

func DoDatalogging(dataloggingRunning *bool, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		if *dataloggingRunning {
			fmt.Println("--- Datalogging: Finish ---")
			return
		}

		fmt.Println("--- Datalogging: Start ---")
		cmd := exec.Command("/home/pi/dev/wt-datalogging/bin/wt-datalogging")
		output, err := cmd.Output()
		if err != nil {
			fmt.Println("Error running datalogging: ", err)
		}

		fmt.Println(string(output))
	}
}