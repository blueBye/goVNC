package main

import (
	"context"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/amitbet/vncproxy/client"
	"github.com/amitbet/vncproxy/common"
	"github.com/amitbet/vncproxy/encodings"
	"github.com/amitbet/vncproxy/logger"
	"github.com/amitbet/vncproxy/recorder"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/remoteconsoles"
	"github.com/gorilla/websocket"
)

func createURLClient(remoteConsole *remoteconsoles.RemoteConsole) url.URL {
	logger.Info("creating a url client")

	// Extract the token from the last 36 characters of the remoteConsole URL
	token := remoteConsole.URL[len(remoteConsole.URL)-36:]

	u := url.URL{
		Scheme:   "ws",
		Host:     os.Getenv("OS_WSOCKHOST"),
		Path:     "/",
		RawQuery: "token=" + token,
	}

	return u
}

func createWSDefaultDialer(remoteConsole *remoteconsoles.RemoteConsole) (*websocket.Conn, error) {
	logger.Info("creating a websocket connection to the noVNC endpoint")

	u := createURLClient(remoteConsole)

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		logger.Error("create ws client error:", err)
		return nil, err
	}

	return conn, nil
}

func createWSWrapper(conn *websocket.Conn) RWC {
	logger.Info("creating a websocket wrapper")

	wsc := RWC{WSC: conn,
		Stream: make(chan byte, 100000),
	}

	return wsc
}

func startWSCWrapper(conn *websocket.Conn, wsc RWC, ctx context.Context) {
	go func() {
		logger.Info("listener goroutine started")

		for {
			// Use a select to handle context cancellation or WebSocket messages
			select {
			case <-ctx.Done():
				logger.Info("listener goroutine received done message, stopping")
				return

			default:
				// // Try to read the WebSocket message, with timeout handling
				_, message, err := conn.ReadMessage()
				n := 100
				if len(message) < 100 {
					n = len(message)
				}
				logger.Info("[generateFBS] message (first 100 bytes):", message[:n])

				if err != nil {
					// If the context was canceled or if there's a WebSocket error, exit
					if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) || err == context.Canceled {
						logger.Info("WebSocket closed or context canceled")
						return
					}

					// Handle WebSocket close errors and stop further operations
					if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
						logger.Error("WebSocket connection unexpectedly closed")
						_ = conn.Close() // Ensure the connection is closed properly
						return
					}

					// Distinguish between timeout and other read errors
					if nErr, ok := err.(net.Error); ok && nErr.Timeout() {
						// Timeout occurred, just continue to check for new messages
						continue
					}

					// Log and break out on any other kind of error
					logger.Warn("wsc read error:", err)
					_ = conn.Close() // Ensure connection closure to prevent further reads
					return
				}

				if len(message) == 0 {
					logger.Info("wswrapper loop received message with size zero")
				}

				logger.Info("writing message to rwc input stream:")
				for _, b := range message {
					wsc.Stream <- b
				}
			}
		}
	}()
}

func createRFBClient(wsc *RWC) (*client.ClientConn, error) {
	logger.Info("connecting to fbs/rfb client")

	var noauth client.ClientAuthNone
	authArr := []client.ClientAuth{&client.PasswordAuth{Password: ""}, &noauth}

	clientConn, err := client.NewClientConn(wsc,
		&client.ClientConfig{
			Auth:      authArr,
			Exclusive: false,
		})

	if err != nil {
		return nil, err
	}

	return clientConn, nil
}

func createRecorderClient(name string) (*recorder.Recorder, error) {
	logger.Info("creating the recorder")
	return recorder.NewRecorder("FBS/" + name + ".fbs")
}

func generateFBS(recordingData RecordingData, name string, remoteConsole *remoteconsoles.RemoteConsole) {
	/** workflow:
	- create a websocket default dialer
	- create a websocket wrapper which implements the interface required by recorder and rfb requester
	- start the websocket wrapper using context to handle closure asynchronously
	- create a RFB client which handles sending the protocol message
	- create a recording client which generates the FBS file
	- connect the recorder to RFB client to record incoming messages (listener)
	- connect a RFB requester to RFB client in order ti send RFB request (listener)
	- start the RFB client
	- wait for n seconds (recordingDuration argument) before returning

	* note: consume method of the RFB requester and recorder will be called by the RBB client when new messages are received
	*/

	// create a websocket default dialer
	conn, connErr := createWSDefaultDialer(remoteConsole)
	if connErr != nil {
		logger.Error("encountered error while connecting to noVNC endpoint:", connErr)
		return
	}

	// create a web socket client
	wsc := createWSWrapper(conn)
	defer func(wsc RWC) {
		_ = wsc.Close()
	}(wsc)

	// start the websocket client wrapper
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startWSCWrapper(conn, wsc, ctx)

	// create a new fbs/rfb client
	rfbClient, rfbErr := createRFBClient(&wsc)
	if rfbErr != nil {
		logger.Error("encountered error while creating RFB client:", rfbErr)
		return
	}
	defer func(rfbClient *client.ClientConn) {
		err := rfbClient.Close()
		if err != nil {
			logger.Error("encountered error while closing RFB client:", err)
		}
	}(rfbClient)

	// create a new recorder
	recClient, err := createRecorderClient(name)
	if err != nil {
		logger.Error("encountered error while creating recorder client:", err)
		return
	}
	defer recClient.Close()

	// add listeners
	rfbClient.Listeners.AddListener(&recorder.RfbRequester{Conn: rfbClient, Name: "Rfb Requester"})
	rfbClient.Listeners.AddListener(recClient)

	// attempt connecting to the noVNC server
	logger.Info("connecting to noVNC server")
	rfbErr = rfbClient.Connect()
	if err != nil {
		logger.Warn("encountered error while connecting to noVNC server:", rfbErr)
		return
	}

	// add encodings to the client after successful connection (see vncproxy/recorder)
	logger.Info("setting encodings")
	encs := []common.IEncoding{
		&encodings.TightEncoding{},
		&encodings.ZLibEncoding{},
		&encodings.CopyRectEncoding{},
		&encodings.PseudoEncoding{int32(common.EncJPEGQualityLevelPseudo8)},
	}
	err = rfbClient.SetEncodings(encs)
	if err != nil {
		logger.Error("SetEncodings failed: ", err)
		return
	}

	// record for n seconds
	time.Sleep(time.Second * time.Duration(recordingData.Duration))
	logger.Info("closing connection after ", recordingData.Duration, " seconds")
}
