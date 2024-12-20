// Copyright (C) 2023-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package asterisk

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/eja/tibula/log"
	"pbx/internal/av"
	"pbx/internal/core"
	"pbx/internal/db"
	"pbx/internal/sys"
)

const recordingTimeout = 30 * 1000

type AgiType struct {
	token     string
	uniqueId  string
	callerId  string
	extension string
	language  string
	request   string
}

func session(conn net.Conn) (err error) {
	const platform = "pbx"
	defer conn.Close()

	reader := bufio.NewReader(conn)
	agi := AgiType{}

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		if strings.HasPrefix(line, "agi_") {
			processHeader(line, &agi)
		} else {
			break
		}
	}

	aiSettings := db.Settings()
	authorized := false
	phone := agi.callerId

	if agi.token == "" {
		if parsedUrl, err := url.Parse(agi.request); err == nil {
			if len(parsedUrl.Path) > 1 {
				agi.token = parsedUrl.Path[1:]
			}
		}
	}

	asteriskToken := aiSettings["asteriskToken"]
	if asteriskToken == "" {
		asteriskToken = sys.Options.AsteriskToken
	}

	asteriskMusic := false
	if sys.Number(aiSettings["asteriskMoh"]) > 0 {
		asteriskMusic = true
	}

	language := aiSettings["language"]
	if agi.language != "en" {
		language = agi.language
	}

	if agi.token != asteriskToken {
		log.Warn(tag, "wrong authorization token")
	} else {
		if sys.Number(aiSettings["userRestricted"]) == 0 {
			authorized = true
		} else {
			user, err := db.UserGet(phone)
			if err == nil && user != nil {
				authorized = true
				language = user["language"]
			}
		}
	}

	if !authorized {
		return fmt.Errorf("user not authorized")
	} else {
		vad := sys.Number(aiSettings["asteriskVad"]) > 0
		talking := false
		monitorFile := fmt.Sprintf("%s/monitor.%s.%s.wav", sys.Options.MediaPath, phone, agi.uniqueId)

		if agi.extension == "h" {
			_, err = core.Chat(platform, phone, "/hangup "+monitorFile, language)
			return
		}

		if _, err = send(conn, "ANSWER"); err != nil {
			return
		}

		mixMonitorTime := time.Now()
		if _, err = send(conn, "EXEC MixMonitor "+monitorFile); err != nil {
			return
		}

		/*
			if vad {
				if _, err = send(conn, "SET MUSIC on"); err != nil {
					return
				}
				start := now()
				if _, err = send(conn, "EXEC WaitForNoise 50,1,2"); err != nil {
					return
				}
				if now()-start <= 2 {
					vad = false
				}
			}
		*/

		dtmf := ""
		if message, err := core.Chat(platform, phone, "/welcome", language); err != nil {
			return err
		} else {
			start := now()
			talking, dtmf, err = play(conn, phone, message, language, vad, mixMonitorTime)
			if err != nil {
				return err
			}
			if vad && talking && (now()-start <= 1) {
				vad = false
				log.Debug(tag, "VAD disabled as line is too noisy")
				talking, dtmf, err = play(conn, phone, message, language, vad, mixMonitorTime)
				if err != nil {
					return err
				}
			}
		}

		for {
			question := ""
			answer := ""
			hangup := false
			ttsLanguage := language

			if dtmf != "" {
				question = dtmf
				dtmf = ""
			} else {
				if question, err = record(conn, phone, language, vad, talking); err != nil {
					return
				}
			}

			if question != "" {
				if asteriskMusic {
					if _, err = send(conn, "SET MUSIC on"); err != nil {
						return
					}
				}
				if answer, err = core.Chat(platform, phone, question, language); err != nil {
					return
				}

				message, tags := core.TagsExtract(answer)
				for _, item := range tags {
					lower := strings.ToLower(item)
					switch lower {
					case "close":
						hangup = true
					case "hangup":
						hangup = true
					}
					if len(lower) == 2 {
						ttsLanguage = lower
						if language == "" {
							language = ttsLanguage
						}
					}
					if strings.HasPrefix(lower, "sip:") {
						if message != "" {
							if talking, _, err = play(conn, phone, message, ttsLanguage, vad, mixMonitorTime); err != nil {
								return
							}
						}
						message = ""
						if _, err = send(conn, fmt.Sprintf("EXEC DIAL PJSIP/%s", item[4:])); err != nil {
							return
						}
						hangup = true
					}
					if strings.HasPrefix(lower, "dial:") {
						if message != "" {
							if talking, _, err = play(conn, phone, message, ttsLanguage, vad, mixMonitorTime); err != nil {
								return
							}
						}
						message = ""
						if _, err = send(conn, fmt.Sprintf("EXEC DIAL %s", item[5:])); err != nil {
							return
						}
						hangup = true
					}
				}
				message, err = core.TagsProcess(platform, language, phone, message, tags)

				if message != "" {
					if talking, dtmf, err = play(conn, phone, message, ttsLanguage, vad, mixMonitorTime); err != nil {
						return
					}
				}

				if hangup {
					_, err = send(conn, "HANGUP")
					return
				}
			}
		}
	}

	return nil
}

func now() int64 {
	return time.Now().Unix()
}

func record(conn net.Conn, phone, language string, vad, talking bool) (string, error) {
	fileName := ""
	silence := 2
	if vad {
		silence = 1
	}
	if !talking {
		if _, err := send(conn, `STREAM FILE beep ""`); err != nil {
			return "", err
		}
	}

	for {
		asteriskFileName := fmt.Sprintf("%s/record.%s.%d", sys.Options.MediaPath, phone, now())
		fileName = asteriskFileName + ".wav16"

		if !talking {
			if _, err := send(conn, "EXEC WaitForNoise 30"); err != nil {
				return "", err
			}
		}
		if msg, err := send(conn, fmt.Sprintf("RECORD FILE %s wav16 # %d 1 s=%d", asteriskFileName, recordingTimeout, silence)); err != nil {
			return msg, err
		}

		if vad {
			talking = false
			vadActivity, err := core.VAD(fileName)
			if err != nil {
				return "", err
			}
			log.Trace(tag, "vad", vadActivity)
			if len(vadActivity) > 0 {
				break
			}
		} else {
			break
		}
	}

	return core.ASR(fileName, language)
}

func play(conn net.Conn, phone string, message string, language string, vad bool, mixMonitorTime time.Time) (talking bool, dtmf string, err error) {
	fileOutputName := fmt.Sprintf("%s/tts.%x.wav", sys.Options.Cache, md5.Sum([]byte(message)))
	fileOutputTmp := fmt.Sprintf("%s/%s.tts.%d", sys.Options.MediaPath, phone, now())
	if _, err = os.Stat(fileOutputName); err != nil {
		if err = core.TTS(message, language, fileOutputTmp); err != nil {
			return
		}
		if err = av.MpegAudioAsterisk(fileOutputTmp, fileOutputName); err != nil {
			return
		}
	} else {
		log.Trace(tag, "using tts transcoded cache for", phone, message)
	}

	if err != nil {
		return
	}
	asteriskFileName := strings.TrimSuffix(fileOutputName, ".wav")

	if vad {
		probeInput, err := av.ProbeAudio(fileOutputName)
		if err != nil {
			return false, "", err
		}
		log.Trace(tag, probeInput)
		playTimeFile := sys.Float(probeInput["duration"])
		playTimeStart := time.Now().UnixNano()
		if msg, err := send(conn, fmt.Sprintf("EXEC BackgroundDetect %s,1,30", asteriskFileName)); err != nil {
			return false, "", err
		} else {
			dtmf = dtmfExtract(msg)
			if dtmf != "" {
				for {
					if msg, err := send(conn, fmt.Sprintf("EXEC WaitExten 1.5")); err != nil {
						return false, "", err
					} else {
						val := dtmfExtract(msg)
						if val == "" {
							break
						} else {
							dtmf += val
						}
					}
				}
			}
		}
		playTimeStop := time.Now().UnixNano()
		playTimeDiff := sys.Float(playTimeStop-playTimeStart) / 1e9
		if playTimeDiff < playTimeFile {
			log.Trace(tag, "talking detected")
			talking = true
		}
	} else {
		_, err = send(conn, fmt.Sprintf("CONTROL STREAM FILE %s # 1000 6 4 5 0", asteriskFileName))
	}

	return
}

func dtmfExtract(msg string) string {
	val := sys.Number(msg[11:])
	if val > 0 {
		return string(val)
	}
	return ""
}

func processHeader(line string, agi *AgiType) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	switch key {
	case "agi_uniqueid":
		agi.uniqueId = value
	case "agi_callerid":
		agi.callerId = value
	case "agi_language":
		agi.language = value
	case "agi_extension":
		agi.extension = value
	case "agi_request":
		agi.request = value
	case "agi_arg_1":
		agi.token = value
	}
}

func send(conn net.Conn, tx string) (rx string, err error) {
	if _, err = conn.Write([]byte(tx + "\n")); err != nil {
		return
	}

	reader := bufio.NewReader(conn)
	if rx, err = reader.ReadString('\n'); err != nil {
		return
	}

	log.Trace(tag, "tx", tx)
	log.Trace(tag, "rx", rx)

	if strings.HasPrefix(rx, "200") {
		return strings.TrimSpace(rx), nil
	} else {
		return rx, fmt.Errorf("%s rx/tx error", tag)
	}
}

func Start() error {
	listener, err := net.Listen("tcp", sys.Options.AsteriskAgi)
	if err != nil {
		return err
	}
	defer listener.Close()

	log.Info(tag, "ready on", sys.Options.AsteriskAgi)
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		go func(conn net.Conn) {
			if err := session(conn); err != nil {
				log.Warn(err)
			}
		}(conn)
	}
}
