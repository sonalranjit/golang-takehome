package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"./activity"
	"./validator"
)

/*Reading files requires checkign most calls for errors. This helper
will streamline our error checks below.*/
func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	args := os.Args[1:]

	if len(args) > 0 {
		//Open file for reading
		f, err := os.Open(args[0])
		check(err)

		scanner := bufio.NewScanner(f)

		for scanner.Scan() {
			data := activity.Activity{}

			//Decode and cast the json data to the Activity type struct
			if err := json.Unmarshal(scanner.Bytes(), &data); err != nil {
				panic(err)
			}
			fmt.Println(validator.ValidateEvent(data))

		}
		if scanner.Err() != nil {
			panic(scanner.Err)
		}
		f.Close()
	} else {
		fmt.Println("Not enough args provided.")
	}

}
