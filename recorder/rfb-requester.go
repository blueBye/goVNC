package recorder

import (
	"time"

	"github.com/amitbet/vncproxy/common"
	"github.com/amitbet/vncproxy/logger"
	"github.com/blueBye/vnc_recorder/client"
)

type RfbRequester struct {
	Conn            *client.ClientConn
	Name            string
	Width           uint16
	Height          uint16
	lastRequestTime time.Time
}

func (p *RfbRequester) Consume(seg *common.RfbSegment) error {

	logger.Debugf("WriteTo.Consume ("+p.Name+"): got segment type=%s", seg.SegmentType)
	switch seg.SegmentType {
	case common.SegmentServerInitMessage:
		serverInitMessage := seg.Message.(*common.ServerInit)
		p.Conn.FrameBufferHeight = serverInitMessage.FBHeight
		p.Conn.FrameBufferWidth = serverInitMessage.FBWidth
		p.Conn.DesktopName = string(serverInitMessage.NameText)
		p.Conn.SetPixelFormat(&serverInitMessage.PixelFormat)
		p.Width = serverInitMessage.FBWidth
		p.Height = serverInitMessage.FBHeight
		p.lastRequestTime = time.Now()
		p.Conn.FramebufferUpdateRequest(false, 0, 0, p.Width, p.Height)

	case common.SegmentMessageStart:
	case common.SegmentRectSeparator:
	case common.SegmentBytes:
	case common.SegmentFullyParsedClientMessage:
	case common.SegmentMessageEnd:
		p.Conn.FramebufferUpdateRequest(true, 0, 0, p.Width, p.Height)
	default:
	}
	return nil
}
