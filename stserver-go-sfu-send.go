package main

import (
	"flag"
	"fmt"

	ilog "github.com/pion/ion-log"
	sdk "github.com/pion/ion-sdk-go"
	"github.com/pion/webrtc/v3"
)

var (
	log = ilog.NewLoggerWithFields(ilog.DebugLevel, "", nil)
)

func main() {
	version := "1.0.3"
	fmt.Println("Version: " + version)

	// parse flag
	var session, addr, videoCodecName, videoBitrate string
	var videoPassthrough bool
	flag.StringVar(&addr, "addr", "localhost:50051", "Ion-sfu grpc addr")
	flag.StringVar(&session, "session", "test session", "join session name")
	flag.StringVar(&videoCodecName, "videoCodecName", "vp8", "WebRTC codec name")
	flag.StringVar(&videoBitrate, "videoBitrate", "3500000", "WebRTC video bitrate")
	flag.BoolVar(&videoPassthrough, "videoPassthrough", false, "Video pathtrough.")

	audioSrc := flag.String("audio-src", "", "Audio src")
	videoSrc := flag.String("video-src", "videotestsrc", "GStreamer video src")
	flag.Parse()

	// add stun servers
	webrtcCfg := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			webrtc.ICEServer{
				URLs: []string{"stun:stun.stunprotocol.org:3478", "stun:stun.l.google.com:19302"},
			},
		},
	}

	config := sdk.Config{
		WebRTC: sdk.WebRTCTransportConfig{
			Configuration: webrtcCfg,
		},
	}
	// new sdk engine
	e := sdk.NewEngine(config)

	// get a client from engine
	c, err := sdk.NewClient(e, addr, "client id")

	var peerConnection *webrtc.PeerConnection = c.GetPubTransport().GetPeerConnection()

	peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		log.Infof("Connection state changed: %s", state)
	})

	if err != nil {
		log.Errorf("client err=%v", err)
		panic(err)
	}

	err = e.AddClient(c)
	if err != nil {
		return
	}

	videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "video/vp8"}, "video", "pion2")
	if err != nil {
		panic(err)
	}

	_, err = peerConnection.AddTrack(videoTrack)
	if err != nil {
		panic(err)
	}

	audioTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "audio/opus"}, "audio", "pion1")
	if err != nil {
		panic(err)
	}
	_, err = peerConnection.AddTrack(audioTrack)
	if err != nil {
		panic(err)
	}

	// client join a session
	err = c.Join(session, nil)

	if err != nil {
		log.Errorf("join err=%v", err)
		panic(err)
	}

	// Start pushing buffers on these tracks
	CreatePipeline("opus", videoPassthrough, "", []*webrtc.TrackLocalStaticSample{audioTrack}, *audioSrc).Start()
	CreatePipeline(videoCodecName, videoPassthrough, videoBitrate, []*webrtc.TrackLocalStaticSample{videoTrack}, *videoSrc).Start()

	select {}
}
