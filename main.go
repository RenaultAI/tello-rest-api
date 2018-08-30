package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/dji/tello"
)

type Command int

const (
	takeOff Command = iota
	land    Command = 1
)

func main() {
	drone := tello.NewDriver("8888")

	commands := make(chan Command)
	errors := make(chan error)

	// This function has to match the work function signature.
	work := func() {
		for cmd := range commands {
			switch cmd {
			case takeOff:
				log.Println("Command takeOff...")
				errors <- drone.TakeOff()

			case land:
			default:
				errors <- fmt.Errorf("Unrecognized command %v", cmd)
			}
		}
	}

	// Start the robot.
	robot := gobot.NewRobot("tello",
		[]gobot.Connection{},
		[]gobot.Device{drone},
		work,
	)

	go func() {
		if err := robot.Start(); err != nil {
			log.Fatal(err)
		}
	}()

	// Start the API server.
	r := gin.Default()

	r.POST("/takeoff", func(c *gin.Context) {
		commands <- takeOff
		if err := <-errors; err != nil {
			log.Printf("Error taking off: %s\n", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "failed",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "success",
		})
	})

	r.POST("/land", func(c *gin.Context) {
		log.Println("landing...")
		if err := drone.Land(); err != nil {
			log.Printf("Error landing: %s\n", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "failed",
			})
			return
		}

		c.JSON(200, gin.H{
			"message": "success",
		})
	})

	address := "0.0.0.0:3000"
	r.Run(address)
}
