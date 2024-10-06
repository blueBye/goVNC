package main

import (
	"github.com/amitbet/vncproxy/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"os"
	"strconv"
	"time"
)

type RecordingData struct {
	ServerID string `json:"server"`
	Duration int    `json:"duration"`
	Count    int    `json:"count"`
}

func recordCreate(c *gin.Context) {
	// read POST request data
	var recording_data RecordingData
	if err := c.BindJSON(&recording_data); err != nil {
		c.IndentedJSON(http.StatusBadRequest, "invalid data")
		return
	}

	// uuid for referencing the bucket containing the videos
	newUUID := uuid.New()
	bucketName := newUUID.String()

	// start the job
	go func() {
		for i := 0; i < recording_data.Count; i++ {
			logger.Info("=================NEW LOOP=====================")
			logger.Info("=================[ ", i, " ]=====================")

			remoteConsole, err := getRemoteConsole(recording_data.ServerID)
			if err != nil {
				c.IndentedJSON(http.StatusCreated, "failed to initiate remote console")
				logger.Warn("failed to initiate remote console:", err)
				return
			}

			name := recording_data.ServerID + "-" + strconv.Itoa(i) + "-" + strconv.FormatInt(time.Now().UTC().UnixNano(), 10)

			// record fbs
			logger.Info("recording " + name)
			generateFBS(recording_data, name, remoteConsole)

			// handle rest of the operations in a routin and move on to recording the next video
			go func(name string) {
				// generate mp4
				filePath := "./FBS/" + name + ".fbs"
				logger.Info("generating mp4 from " + filePath)
				_ = generateVideo(filePath, float64(recording_data.Duration))

				// fix video duration due to lag in processing
				time.Sleep(time.Second)
				currentDuration, _ := getVideoDuration(filePath + ".mp4")
				logger.Info("current duration:", currentDuration)
				_ = fixSpeed(filePath, currentDuration, float64(recording_data.Duration))

				// cleanup
				_ = os.Remove(filePath)
				_ = os.Remove(filePath + ".mp4")

				// upload to minio
				filePath += "-fixed.mp4"
				logger.Info("uploading", filePath, "to bucket", bucketName)
				upload(filePath, bucketName)
				_ = os.Remove(filePath)
			}(name)
		}
	}()

	c.IndentedJSON(http.StatusCreated, gin.H{"message": "job started", "bucket": bucketName})
}

func main() {
	// set loggin level and load envoronment variables from .evn file
	logger.SetLogLevel("info")

	router := gin.Default()
	router.POST("/record", recordCreate)
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	router.Run("0.0.0.0:8080")
}
