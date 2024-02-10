package main

import (
	"errors"
	"net/url"
	"os"
	"time"

	vnc "github.com/amitbet/vnc2video"
	"github.com/amitbet/vnc2video/encoders"
	"github.com/amitbet/vncproxy/common"
	"github.com/amitbet/vncproxy/encodings"
	"github.com/amitbet/vncproxy/logger"
	"github.com/blueBye/vnc_recorder/client"
	"github.com/blueBye/vnc_recorder/recorder"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/remoteconsoles"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

func getRemoteConsole() (*remoteconsoles.RemoteConsole, error) {
	opts := gophercloud.AuthOptions{
		IdentityEndpoint: os.Getenv("OS_IDENTITYENDPOINT"),
		Username:         os.Getenv("OS_USERNAME"),
		Password:         os.Getenv("OS_PASSWORD"),
		TenantID:         os.Getenv("OS_TENANTID"),
		DomainID:         os.Getenv("OS_FOMAINID"),
	}

	provider, _ := openstack.AuthenticatedClient(opts)
	compute, _ := openstack.NewComputeV2(provider, gophercloud.EndpointOpts{Region: "RegionOne"})
	compute.Microversion = "2.6"

	createOpts := remoteconsoles.CreateOpts{
		Protocol: remoteconsoles.ConsoleProtocolVNC,
		Type:     remoteconsoles.ConsoleTypeNoVNC,
	}
	serverID := os.Getenv("OS_SERVERID")

	remoteConsole, err := remoteconsoles.Create(compute, serverID, createOpts).Extract()
	if err != nil {
		logger.Error("server url not found:", err)
		return nil, errors.New("failed to ")
	}

	return remoteConsole, nil
}

func generateVideo(input string, output string) {
	framerate := 20
	speedupFactor := 1.2
	fastFramerate := int(float64(framerate) * speedupFactor)

	video_encs := []vnc.Encoding{
		&vnc.RawEncoding{},
		&vnc.ZLibEncoding{},
		&vnc.TightEncoding{},
		&vnc.CopyRectEncoding{},
		&vnc.ZRLEEncoding{},
	}

	fbs, err := vnc.NewFbsConn(input, video_encs)
	if err != nil {
		logger.Error("failed to open fbs reader:", err)
	}

	//launch video encoding process:
	vcodec := &encoders.X264ImageEncoder{FFMpegBinPath: "ffmpeg", Framerate: framerate}
	go vcodec.Run(output)

	screenImage := vnc.NewVncCanvas(int(fbs.Width()), int(fbs.Height()))
	screenImage.DrawCursor = false

	for _, enc := range video_encs {
		myRenderer, ok := enc.(vnc.Renderer)

		if ok {
			myRenderer.SetTargetImage(screenImage)
		}
	}

	go func() {
		frameMillis := (1000.0 / float64(fastFramerate)) - 1
		frameDuration := time.Duration(frameMillis * float64(time.Millisecond))

		for {
			timeStart := time.Now()

			vcodec.Encode(screenImage.Image)
			timeTarget := timeStart.Add(frameDuration)
			timeLeft := time.Until(timeTarget)
			if timeLeft > 0 {
				time.Sleep(timeLeft)
			}
		}
	}()

	msgReader := vnc.NewFBSPlayHelper(fbs)
	//loop over all messages, feed images to video codec:
	for {
		_, err := msgReader.ReadFbsMessage(true, speedupFactor)
		vcodec.Encode(screenImage.Image)
		if err != nil {
			os.Exit(-1)
		}
	}
}

func main() {
	logger.SetLogLevel("warn")
	state := "recording"

	godotenv.Load(".env")

	remoteConsole, _ := getRemoteConsole()

	u := url.URL{
		Scheme:   "ws",
		Host:     os.Getenv("OS_WSOCKHOST"),
		Path:     "/",
		RawQuery: "token=" + remoteConsole.URL[len(remoteConsole.URL)-36:]}

	// Create a new websocket connection to the noVNC endpoint
	logger.Info("creating a websocker connection to the noVNC endpoint")
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		logger.Error("dial error:", err)
		return
	}

	// define a new Read/Write consumer (implements net.Dial interface with websocket)
	logger.Info("creating a read/write consumer")
	var noauth client.ClientAuthNone
	authArr := []client.ClientAuth{&client.PasswordAuth{Password: ""}, &noauth}
	wsc := client.RWC{WSC: conn,
		Buffer:      make(chan []byte, 1),
		InputStream: make(chan []byte, 100)}

	defer wsc.Close()

	// websocket message handler goroutine (incomming messages): writes messages to wsc buffer
	go func() {
		logger.Info("listener goroutine started")
		for {
			if state != "recording" {
				return
			}

			_, message, err := conn.ReadMessage()
			if err != nil {
				logger.Error("wsc read error:", err)
				return
			}
			wsc.InputStream <- []byte(message)
		}
	}()

	// create a new fbs/rfb client (noVNC client)
	logger.Info("connecting to fbs/rfb client")
	clientConn, _ := client.NewClientConn(wsc,
		&client.ClientConfig{
			Auth:      authArr,
			Exclusive: true,
		})

	// create a new recorder to save the rfb traffic as a fbs file
	logger.Info("creating the recorder")
	rec, err := recorder.NewRecorder("autorec.fbs")
	if err != nil {
		logger.Error("error creating recorder:", err)
		return
	}

	defer rec.Close()

	// add listeners to the clientConn for recorder (see vncproxy/recorder)
	clientConn.Listeners.AddListener(&recorder.RfbRequester{Conn: clientConn, Name: "Rfb Requester"})
	clientConn.Listeners.AddListener(rec)

	// attempt connecting to the noVNC server
	logger.Info("connecting to noVNC server")
	clientConn.Connect()
	if err != nil {
		logger.Error("error while connecting to noVNC server:", err)
		return
	}

	// add encodings to the client after successful connection (see vncproxy/recorder)
	logger.Info("setting encodings")
	encs := []common.IEncoding{
		// &encodings.TightEncoding{},
		&encodings.ZLibEncoding{},
	}
	clientConn.SetEncodings(encs)

	// record for 10 seconds
	time.Sleep(time.Second * 20)
	state = "converting"
	logger.Info("closing connection")
	generateVideo("autorec.fbs", "output.mp4")
}
