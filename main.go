package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/amitbet/vncproxy/logger"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type RecordingData struct {
	ServerID string `json:"server"`
	Duration int    `json:"duration"`
}

func recordCreate(c *gin.Context) {
	// read POST request data
	var recording_data RecordingData
	if err := c.BindJSON(&recording_data); err != nil {
		c.IndentedJSON(http.StatusBadRequest, "invalid data")
		return
	}

	remoteConsole, err := getRemoteConsole(recording_data.ServerID)
	if err != nil {
		c.IndentedJSON(http.StatusCreated, "failed to initiate remote console")
		logger.Warn("failed to initiate remote console:", err)
		return
	}

	// start the job
	go func() {
		timestamp := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
		name := recording_data.ServerID + "_" + timestamp

		// record fbs
		generateFBS(recording_data, name, remoteConsole)

		// generate mp4
		if err := generateVideo(name + ".fbs"); err != nil {
			logger.Error("encountered error while generating mp4 file:", err)
		}

		// upload to minio
		upload(name + ".fbs.mp4")

		// notify the django (doesn't know the timestamp to access file)
	}()

	c.IndentedJSON(http.StatusCreated, "job started")

}

func main() {
	// set loggin level and load envoronment variables from .evn file
	logger.SetLogLevel("warn")
	godotenv.Load(".env")

	router := gin.Default()
	router.POST("/record", recordCreate)

	router.Run("localhost:8080")
}
