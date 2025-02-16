package interactive

// func filterInput(r rune) (rune, bool) {
// 	switch r {
// 	// block CtrlZ feature
// 	case readline.CharCtrlZ:
// 		return r, false
// 	}
// 	return r, true
// }

// func Process(cfg config.Config, wg *sync.WaitGroup, rl *readline.Instance) {

// 	for {
// 		line, err := rl.Readline()
// 		if err == readline.ErrInterrupt {
// 			if len(line) == 0 {
// 				break
// 			} else {
// 				continue
// 			}
// 		} else if err == io.EOF {
// 			break
// 		}

// 		line = strings.TrimSpace(line)
// 		if line == "" {
// 			continue
// 		}

// 		switch {
// 		case line == "/ping":

// 		case line == "/help":
// 			usage(l.Stderr())
// 			continue
// 		case strings.HasPrefix(line, "/exit"):
// 			return
// 		case strings.HasPrefix(line, "/connect"):
// 			line = strings.TrimSpace(line[8:])
// 			conn, _, err := ws.Connect(cfg)
// 			if err != nil {
// 				log.Println("connect err,", err)
// 			} else {
// 				go ws.ReadMessages(cfg, conn, wg, l)
// 				log.Println("Connected to ", line)
// 			}
// 		default:
// 			// ws.WriteToServer(WSConn, line)
// 		}
// 	}
// }
