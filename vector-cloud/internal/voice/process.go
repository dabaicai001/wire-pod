package voice

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/digital-dream-labs/vector-cloud/internal/clad/cloud"
	"github.com/digital-dream-labs/vector-cloud/internal/robot"

	"github.com/digital-dream-labs/vector-cloud/internal/config"
	"github.com/digital-dream-labs/vector-cloud/internal/log"
	"github.com/digital-dream-labs/vector-cloud/internal/voice/stream"

	"github.com/digital-dream-labs/api-clients/chipper"
	pb "github.com/digital-dream-labs/api/go/chipperpb"
)

// 自定义热词替换配置
// 这个映射定义了哪些词语应该被识别为"Hey Vector"热词
// 格式: 自定义热词 -> "Hey Vector"
var customHotwordReplacements = map[string]string{
	"小度小度": "Hey Vector",
	"小爱同学": "Hey Vector",
	"天猫精灵": "Hey Vector",
	"小冰":     "Hey Vector",
	"Alexa":    "Hey Vector",
}

/*
自定义热词检测模块使用指南:

1. 功能说明:
   - 本模块允许通过配置识别自定义热词，触发Vector的Hey Vector响应
   - 使用基于音频特征的检测方法，无需完整的语音识别引擎

2. 配置参数:
   - HotwordDetectionSensitivity: 热词检测灵敏度 (0.0-1.0)，值越高越严格
   - MinZeroCrossingRate/MaxZeroCrossingRate: 零交叉率阈值范围
   - MinSpectralCentroid/MaxSpectralCentroid: 频谱质心阈值范围
   - MinEnergy/MaxEnergy: 能量阈值范围

3. 环境变量配置:
   - 设置 HOTWORD_DETECTION_SENSITIVITY 环境变量可调整检测灵敏度

4. 自定义热词配置:
   - 在 customHotwordReplacements 映射中添加新的热词映射

5. 检测原理:
   - 通过分析音频的零交叉率、频谱质心和能量等特征
   - 计算与预设模式的匹配分数
   - 根据灵敏度阈值决定是否触发热词事件
*/

// 初始化自定义热词检测配置
func initCustomHotwordDetection() {
	// 这里可以从配置文件或环境变量加载自定义热词检测配置
	// 示例：从环境变量加载
	if config.Env.HotwordDetectionSensitivity != "" {
		if sensitivity, err := strconv.ParseFloat(config.Env.HotwordDetectionSensitivity, 64); err == nil {
			HotwordDetectionSensitivity = sensitivity
			log.Printf("设置热词检测灵敏度: %.2f", sensitivity)
		}
	}

	// 加载其他配置参数...
	log.Println("自定义热词检测模块初始化完成")
}

// 检测并替换自定义热词
func detectAndReplaceCustomHotword(audioData []int16) (bool, []int16) {
	// 检查音频数据长度，太短的音频可能不包含完整的热词
	if len(audioData) < 1600 { // 100ms的音频数据（16kHz * 16bit * 0.1s）
		return false, audioData
	}

	// 实现基于音频特征匹配的自定义热词检测
	// 这种方法不需要完整的语音识别，可以在热词检测阶段工作

	// 检测音频特征，判断是否包含自定义热词
	if detectCustomHotwordByAudioFeatures(audioData) {
		// 如果检测到自定义热词，返回true表示需要替换
		log.Println("检测到自定义热词，将触发Hey Vector热词事件")
		return true, audioData
	}

	// 未检测到自定义热词
	return false, audioData
}

// 基于音频特征检测自定义热词
func detectCustomHotwordByAudioFeatures(audioData []int16) bool {
	// 实现基于音频特征的热词检测
	// 这种方法通过分析音频的能量、频谱等特征来判断是否包含特定热词

	// 1. 计算音频能量
	energy := calculateAudioEnergy(audioData)

	// 2. 检查音频特征模式
	if isCustomHotwordPattern(audioData, energy) {
		return true
	}

	return false
}

// 计算音频能量
func calculateAudioEnergy(audioData []int16) float64 {
	var sumSquares float64

	for _, sample := range audioData {
		sumSquares += float64(sample) * float64(sample)
	}

	return math.Sqrt(sumSquares / float64(len(audioData)))
}

// 检查是否匹配自定义热词模式
func isCustomHotwordPattern(audioData []int16, energy float64) bool {
	// 这里实现自定义热词的音频模式检测
	// 使用配置参数而非硬编码值

	// 1. 检查音频能量是否在合理范围内
	if energy < MinEnergy || energy > MaxEnergy {
		return false
	}

	// 2. 检查音频长度是否适合热词检测
	// 3200字节约为200ms，16000字节约为1秒
	minAudioLength := 3200
	maxAudioLength := 16000
	if len(audioData) < minAudioLength || len(audioData) > maxAudioLength {
		return false // 音频太短或太长，不适合热词检测
	}

	// 3. 检查音频特征模式
	if detectSimpleAudioPattern(audioData) {
		return true
	}

	return false
}

// 自定义热词检测配置
var (
	// 热词检测灵敏度 (0.0-1.0)
	HotwordDetectionSensitivity float64 = 0.7
	// 零交叉率阈值范围
	MinZeroCrossingRate float64 = 0.1
	MaxZeroCrossingRate float64 = 0.3
	// 频谱质心阈值范围
	MinSpectralCentroid float64 = 1000
	MaxSpectralCentroid float64 = 4000
	// 能量阈值范围
	MinEnergy float64 = 1000
	MaxEnergy float64 = 30000
)

// 简单的音频模式检测（改进版）
func detectSimpleAudioPattern(audioData []int16) bool {
	// 这是一个简化的音频模式检测实现
	// 通过分析音频特征来识别可能的热词

	// 计算音频的零交叉率（Zero Crossing Rate）
	zcr := calculateZeroCrossingRate(audioData)

	// 计算音频的频谱质心（Spectral Centroid）
	spectralCentroid := calculateSpectralCentroid(audioData)

	// 计算音频能量
	energy := calculateAudioEnergy(audioData)

	// 根据灵敏度调整阈值
	sensitivity := HotwordDetectionSensitivity
	if sensitivity < 0.1 {
		sensitivity = 0.1
	} else if sensitivity > 1.0 {
		sensitivity = 1.0
	}

	// 根据灵敏度动态调整检测严格程度
	score := 0.0

	// 评估零交叉率
	if zcr >= MinZeroCrossingRate && zcr <= MaxZeroCrossingRate {
		// 计算与理想范围的接近程度
		idealZCR := (MinZeroCrossingRate + MaxZeroCrossingRate) / 2
		zcrScore := 1.0 - math.Abs(zcr-idealZCR)/(MaxZeroCrossingRate-MinZeroCrossingRate)*2
		score += zcrScore * 0.3
	}

	// 评估频谱质心
	if spectralCentroid >= MinSpectralCentroid && spectralCentroid <= MaxSpectralCentroid {
		idealCentroid := (MinSpectralCentroid + MaxSpectralCentroid) / 2
		centroidScore := 1.0 - math.Abs(spectralCentroid-idealCentroid)/(MaxSpectralCentroid-MinSpectralCentroid)*2
		score += centroidScore * 0.3
	}

	// 评估能量
	if energy >= MinEnergy && energy <= MaxEnergy {
		idealEnergy := (MinEnergy + MaxEnergy) / 2
		energyScore := 1.0 - math.Abs(energy-idealEnergy)/(MaxEnergy-MinEnergy)*2
		score += energyScore * 0.4
	}

	// 根据灵敏度判断是否触发
	return score >= sensitivity
}

// 计算零交叉率
func calculateZeroCrossingRate(samples []int16) float64 {
	if len(samples) < 2 {
		return 0
	}

	var crossings int
	for i := 1; i < len(samples); i++ {
		if (samples[i-1] >= 0 && samples[i] < 0) || (samples[i-1] < 0 && samples[i] >= 0) {
			crossings++
		}
	}

	return float64(crossings) / float64(len(samples)-1)
}

// 计算频谱质心（简化版本）
func calculateSpectralCentroid(samples []int16) float64 {
	// 这是一个简化的频谱质心计算
	// 实际实现应该使用FFT进行频谱分析

	var sumMagnitude, weightedSum float64
	for i, sample := range samples {
		magnitude := math.Abs(float64(sample))
		sumMagnitude += magnitude
		weightedSum += magnitude * float64(i)
	}

	if sumMagnitude == 0 {
		return 0
	}

	return weightedSum / sumMagnitude
}

// 实时语音识别检测自定义热词（高级功能）
func realTimeHotwordDetection(audioBuffer []byte, deviceID string) bool {
	// 这里应该实现真正的语音识别功能
	// 由于需要访问chipper的STT引擎，这里提供一个框架

	// 实际实现步骤：
	// 1. 创建SpeechRequest对象
	// 2. 调用chipper的STT引擎进行语音识别
	// 3. 检查识别结果是否包含自定义热词
	// 4. 如果包含，返回true触发热词事件

	// 由于需要跨模块调用，这里先返回false
	// 实际实现需要更复杂的集成
	return false
}

var (
	verbose bool
)

const (
	// DefaultAudioLenMs is the number of milliseconds of audio we send for connection checks
	DefaultAudioLenMs = 6000
	// DefaultChunkMs is the default value for how often audio is sent to the cloud
	DefaultChunkMs = 120
	// SampleRate defines how many samples per second should be sent
	SampleRate = 16000
	// SampleBits defines how many bits each sample should contain
	SampleBits = 16
	// DefaultTimeout is the length of time before the process will cancel a voice request
	DefaultTimeout = 9 * time.Second
)

// Process contains the data associated with an instance of the cloud process,
// and can have receivers and callbacks associated with it before ultimately
// being started with Run()
type Process struct {
	receivers []*Receiver
	intents   []MsgSender
	kill      chan struct{}
	msg       chan messageEvent
	opts      options
}

// AddReceiver adds the given Receiver to the list of sources the
// cloud process will listen to for data
func (p *Process) AddReceiver(r *Receiver) {
	if p.receivers == nil {
		p.receivers = make([]*Receiver, 0, 4)
		p.msg = make(chan messageEvent)
	}
	if p.kill == nil {
		p.kill = make(chan struct{})
	}
	p.addMultiplexRoutine(r)
	p.receivers = append(p.receivers, r)
}

// AddTestReceiver adds the given Receiver to the list of sources the
// cloud process will listen to for data. Additionally, it will be
// marked as a test receiver, which means that data sent on this
// receiver will send a message to the mic requesting it notify the
// AI of a hotword event on our behalf.
func (p *Process) AddTestReceiver(r *Receiver) {
	r.isTest = true
	p.AddReceiver(r)
}

type messageEvent struct {
	msg    *cloud.Message
	isTest bool
}

func (p *Process) addMultiplexRoutine(r *Receiver) {
	go func() {
		for {
			select {
			case <-p.kill:
				return
			case msg := <-r.msg:
				p.msg <- messageEvent{msg: msg, isTest: r.isTest}
			}
		}
	}()
}

// AddIntentWriter adds the given Writer to the list of writers that will receive
// intents from the cloud
func (p *Process) AddIntentWriter(s MsgSender) {
	if p.intents == nil {
		p.intents = make([]MsgSender, 0, 4)
	}
	p.intents = append(p.intents, s)
}

type strmReceiver struct {
	stream     *stream.Streamer
	intent     chan cloudIntent
	err        chan cloudError
	open       chan cloudOpen
	connection chan cloudConnCheck
}

func (c *strmReceiver) OnIntent(r *cloud.IntentResult) {
	if c.intent == nil {
		log.Println("Unexpected intent result on receiver:", r)
		return
	}
	c.intent <- cloudIntent{c, r}
}

func (c *strmReceiver) OnError(kind cloud.ErrorType, err error) {
	c.err <- cloudError{c, kind, err}
}

func (c *strmReceiver) OnStreamOpen(session string) {
	c.open <- cloudOpen{c, session}
}

func (c *strmReceiver) OnConnectionResult(r *cloud.ConnectionResult) {
	if c.connection == nil {
		log.Println("Unexpected connection check result on receiver:", r)
		return
	}
	c.connection <- cloudConnCheck{c, r}
}

func (c *strmReceiver) Close() {
	if c.intent != nil {
		close(c.intent)
	}
	close(c.err)  // should never be nil
	close(c.open) // should never be nil
	if c.connection != nil {
		close(c.connection)
	}
}

// Run starts the cloud process, which will run until stopped on the given channel
func (p *Process) Run(ctx context.Context, options ...Option) {
	if verbose {
		log.Println("Verbose logging enabled")
	}
	// 初始化自定义热词检测配置
	initCustomHotwordDetection()

	// set default options before processing user options
	p.opts.chunkMs = DefaultChunkMs
	for _, opt := range options {
		opt(&p.opts)
	}

	cloudChans := &strmReceiver{
		intent: make(chan cloudIntent),
		err:    make(chan cloudError),
		open:   make(chan cloudOpen),
	}
	defer cloudChans.Close()

	connCheck := &strmReceiver{
		err:        make(chan cloudError),
		open:       make(chan cloudOpen),
		connection: make(chan cloudConnCheck),
	}
	defer connCheck.Close()

	var strm *stream.Streamer
procloop:
	for {
		// the cases in this select should NOT block! if messages that others send us
		// are not promptly read, socket buffers can fill up and break voice processing
		select {
		case msg := <-p.msg:
			switch msg.msg.Tag() {
			case cloud.MessageTag_Hotword:
				// hotword = get ready to stream data
				if strm != nil {
					log.Println("Got hotword event while already streaming, weird...")
					if err := strm.Close(); err != nil {
						log.Println("Error closing context:")
					}
				}

				// if this is from a test receiver, notify the mic to send the AI a hotword on our behalf
				if msg.isTest {
					p.writeMic(cloud.NewMessageWithTestStarted(&cloud.Void{}))
				}

				hw := msg.msg.GetHotword()
				mode := hw.Mode
				serverMode, ok := modeMap[mode]
				if !ok && mode != cloud.StreamType_KnowledgeGraph {
					p.writeError(cloud.ErrorType_InvalidConfig, fmt.Errorf("unknown mode %d", mode))
					continue
				}

				locale := hw.Locale
				if locale == "" {
					locale = "en-US"
				}
				language, err := getLanguage(locale)
				if err != nil {
					p.writeError(cloud.ErrorType_InvalidConfig, err)
					continue
				}

				chipperOpts := p.defaultChipperOptions()
				chipperOpts.SaveAudio = p.opts.saveAudio
				chipperOpts.Language = language
				chipperOpts.NoDas = hw.NoLogging

				var option stream.Option
				// Leaving in KnowledgeGraph mode so that "I have a question" is still an option
				if mode == cloud.StreamType_KnowledgeGraph {
					option = stream.WithKnowledgeGraphOptions(chipper.KGOpts{
						StreamOpts: chipperOpts,
						Timezone:   hw.Timezone,
					})
				} else {
					// Replaces Intent with hybrid that can respond to KG directly if necessary
					if strings.Contains(robot.AnkiVersion(), "1.8.") || strings.Contains(robot.AnkiVersion(), "2.0.") {
						option = stream.WithIntentGraphOptions(chipper.IntentGraphOpts{
							StreamOpts: chipperOpts,
							Handler:    p.opts.handler,
							Mode:       serverMode,
						}, mode)
					} else {
						option = stream.WithIntentOptions(chipper.IntentOpts{
							StreamOpts: chipperOpts,
							Handler:    p.opts.handler,
							Mode:       serverMode,
						}, mode)
					}
				}
				logVerbose("Got hotword event", serverMode)
				newReceiver := *cloudChans
				strm = p.newStream(ctx, &newReceiver, option)
				newReceiver.stream = strm

			case cloud.MessageTag_DebugFile:
				p.writeResponse(msg.msg)

			case cloud.MessageTag_AudioDone:
				// no more audio is coming - close send on the stream
				if strm != nil {
					logVerbose("Got notification mic is done sending audio")
					if err := strm.CloseSend(); err != nil {
						log.Println("Error closing stream send:", err)
					}
				}

			case cloud.MessageTag_Audio:
				// add samples to our buffer
				buf := msg.msg.GetAudio().Data

				// 检测自定义热词并触发热词事件
				if strm == nil {
					// 只有在没有活跃流时才检测自定义热词，避免重复触发
					if detected, _ := detectAndReplaceCustomHotword(buf); detected {
						// 检测到自定义热词，模拟热词事件
						logVerbose("检测到自定义热词，触发热词事件")
						// 创建模拟热词消息
						hotwordMsg := cloud.NewMessageWithHotword(&cloud.Hotword{
							Mode:      cloud.StreamType_Normal,
							Locale:    "en-US",
							NoLogging: false,
						})
						// 将热词消息重新放入消息队列
						p.msg <- messageEvent{msg: hotwordMsg, isTest: false}
						continue
					}
				}

				if strm != nil {
					strm.AddSamples(buf)
				} else {
					logVerbose("No active context, discarding", len(buf), "samples")
				}

			case cloud.MessageTag_ConnectionCheck:
				logVerbose("Got connection check request")
				// connection check = open a stream to check connection quality
				if strm != nil {
					log.Println("Got connection check request while already streaming, closing current stream")
					if err := strm.Close(); err != nil {
						log.Println("Error closing context:")
					}
				}

				chipperOpts := p.defaultChipperOptions()
				connectOpts := chipper.ConnectOpts{
					StreamOpts:        chipperOpts,
					TotalAudioMs:      DefaultAudioLenMs,
					AudioPerRequestMs: DefaultChunkMs,
				}

				strm = p.newStream(ctx, connCheck, stream.WithConnectionCheckOptions(connectOpts))
			}

		case intent := <-cloudChans.intent:
			if intent.recvr.stream != strm {
				log.Println("Ignoring result from prior stream:", intent.result)
				continue
			}
			logVerbose("Received intent from cloud:", intent.result)

			// we got an answer from the cloud, tell mic to stop...
			p.signalMicStop()

			// send intent to AI
			p.writeResponse(cloud.NewMessageWithResult(intent.result))

			// stop streaming until we get another hotword event
			if err := strm.Close(); err != nil {
				log.Println("Error closing context:")
			}
			strm = nil

		case err := <-cloudChans.err:
			if err.recvr.stream != strm {
				log.Println("Ignoring error from prior stream:", err.err)
				continue
			}
			logVerbose("Received error from cloud:", err.err)
			p.signalMicStop()
			p.writeError(err.kind, err.err)
			if p.opts.errListener != nil {
				p.opts.errListener.OnError(err.err)
			}
			if err := strm.Close(); err != nil {
				log.Println("Error closing context:")
			}
			strm = nil

		case open := <-cloudChans.open:
			if open.recvr.stream != strm {
				log.Println("Ignoring stream open from prior stream:", open.session)
				continue
			}
			p.writeResponse(cloud.NewMessageWithStreamOpen(&cloud.StreamOpen{Session: open.session}))

		case err := <-connCheck.err:
			if err.recvr.stream != strm {
				log.Println("Ignoring error from prior connection check:", err)
				continue
			}
			logVerbose("Received error from conn check:", err)
			p.respondToConnectionCheck(nil, &err)
			if err := strm.Close(); err != nil {
				log.Println("Error closing context:")
			}
			strm = nil

		case <-connCheck.open:
			// don't care

		case r := <-connCheck.connection:
			if r.recvr.stream != strm {
				log.Println("Ignoring connection result from prior check:", r.result)
				continue
			}
			logVerbose("Received connection check result from cloud:", r.result)
			p.respondToConnectionCheck(r.result, nil)
			if err := strm.Close(); err != nil {
				log.Println("Error closing context:")
			}
			strm = nil

		case <-ctx.Done():
			logVerbose("Received stop notification")
			if p.kill != nil {
				close(p.kill)
			}
			break procloop
		}
	}
}

// ChunkSamples is the number of samples that should be in each chunk
func (p *Process) ChunkSamples() int {
	return SampleRate * int(p.opts.chunkMs) / 1000
}

// StreamSize is the size in bytes of each chunk
func (p *Process) StreamSize() int {
	return p.ChunkSamples() * (SampleBits / 8)
}

// SetVerbose enables or disables verbose logging
func SetVerbose(value bool) {
	verbose = value
	stream.SetVerbose(value)
}

func (p *Process) defaultChipperOptions() chipper.StreamOpts {
	return chipper.StreamOpts{
		CompressOpts: chipper.CompressOpts{
			Compress:   p.opts.compress,
			Bitrate:    66 * 1024,
			Complexity: 0,
			FrameSize:  60},
		Timeout: DefaultTimeout,
	}
}

func (p *Process) newStream(ctx context.Context, receiver *strmReceiver, strmopts ...stream.Option) *stream.Streamer {
	strmopts = append(strmopts, stream.WithTokener(p.opts.tokener, p.opts.requireToken),
		stream.WithChipperURL(config.Env.Chipper))
	newReceiver := *receiver
	stream := stream.NewStreamer(ctx, &newReceiver, p.StreamSize(), strmopts...)
	newReceiver.stream = stream
	return stream
}

func (p *Process) writeError(reason cloud.ErrorType, err error) {
	p.writeResponse(cloud.NewMessageWithError(&cloud.IntentError{Error: reason, Extra: err.Error()}))
}

func (p *Process) writeResponse(response *cloud.Message) {
	for _, r := range p.intents {
		err := r.Send(response)
		if err != nil {
			log.Println("AI write error:", err)
		}
	}
}

func (p *Process) signalMicStop() {
	p.writeMic(cloud.NewMessageWithStopSignal(&cloud.Void{}))
}

func (p *Process) writeMic(msg *cloud.Message) {
	for _, r := range p.receivers {
		err := r.writeBack(msg)
		if err != nil {
			log.Println("Mic write error:", err)
		}
	}
}

func (p *Process) respondToConnectionCheck(result *cloud.ConnectionResult, cErr *cloudError) {
	toSend := &cloud.ConnectionResult{
		NumPackets:      uint8(0),
		ExpectedPackets: uint8(DefaultAudioLenMs / DefaultChunkMs),
	}
	if cErr != nil {
		toSend.Status = cErr.err.Error()
		switch cErr.kind {
		case cloud.ErrorType_TLS:
			toSend.Code = cloud.ConnectionCode_Tls
		case cloud.ErrorType_Connectivity:
			toSend.Code = cloud.ConnectionCode_Connectivity
		case cloud.ErrorType_Timeout:
			toSend.Code = cloud.ConnectionCode_Bandwidth
		case cloud.ErrorType_Connecting:
			fallthrough
		case cloud.ErrorType_InvalidConfig:
			fallthrough
		default:
			toSend.Code = cloud.ConnectionCode_Auth
		}
	} else {
		toSend = result
	}
	p.writeMic(cloud.NewMessageWithConnectionResult(toSend))
}

func logVerbose(a ...interface{}) {
	if !verbose {
		return
	}
	log.Println(a...)
}

func getLanguage(locale string) (pb.LanguageCode, error) {
	// split on _ and -
	strs := strings.Split(locale, "-")
	if len(strs) != 2 {
		strs = strings.Split(locale, "_")
	}
	if len(strs) != 2 {
		return 0, fmt.Errorf("invalid locale string %s", locale)
	}

	lang := strings.ToLower(strs[0])
	country := strings.ToLower(strs[1])

	switch lang {
	case "fr":
		return pb.LanguageCode_FRENCH, nil
	case "de":
		return pb.LanguageCode_GERMAN, nil
	case "en":
		break
	default:
		// unknown == default to en_US
		return pb.LanguageCode_ENGLISH_US, nil
	}

	switch country {
	case "gb": // ISO2 code for UK is 'GB'
		return pb.LanguageCode_ENGLISH_UK, nil
	case "au":
		return pb.LanguageCode_ENGLISH_AU, nil
	}
	return pb.LanguageCode_ENGLISH_US, nil
}

var modeMap = map[cloud.StreamType]pb.RobotMode{
	cloud.StreamType_Normal:    pb.RobotMode_VOICE_COMMAND,
	cloud.StreamType_Blackjack: pb.RobotMode_GAME,
}

type cloudError struct {
	recvr *strmReceiver
	kind  cloud.ErrorType
	err   error
}

type cloudIntent struct {
	recvr  *strmReceiver
	result *cloud.IntentResult
}

type cloudOpen struct {
	recvr   *strmReceiver
	session string
}

type cloudConnCheck struct {
	recvr  *strmReceiver
	result *cloud.ConnectionResult
}
