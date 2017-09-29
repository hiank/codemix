package codemix

import (
	"time"
	"math/rand"
	"bytes"
)

func mixH(h []byte) (rlt []byte, list []*FuncInfo) {

	list = make([]*FuncInfo, 0, 100)
	

	tmp := cleanBuf(h)
	// fmt.Printf("tmp is :%s\n", string(tmp))
	tmpSlice, hSlice := tmp, h
	tmpblockArr, blockArr, classArr := make([][]byte, 0, 3), make([][]byte, 0, 3), [][]byte{}
	for {
		
		l, r := nextClassSite(tmpSlice)
		if l == -1 {
			blockArr = append(blockArr, hSlice)
			// rlt = append(rlt, hSlice...)
			break
		}

		cn := className(hSlice[l:r])
		classArr = append(classArr, cn)
		// classArr = append(classArr, classInfo())

		blockArr = append(blockArr, hSlice[:l])
		blockArr = append(blockArr, hSlice[l:r])
		tmpblockArr = append(tmpblockArr, tmpSlice[:l])
		tmpblockArr = append(tmpblockArr, tmpSlice[l:r])
		
		tmpSlice, hSlice = tmpSlice[r:], hSlice[r:]
	}

	rand.Seed(time.Now().UnixNano())	
	for i, v := range classArr {

		blockIdx := i*2 + 1
		tmpblock := tmpblockArr[blockIdx]

		num := 20 + rand.Intn(80)
		fi := make([]*FuncInfo, num)
		randFunc(fi)
		istatic := bytes.Contains(tmpblock, []byte("static "))
		for x := range fi {

			fi[x].istatic = istatic
			fi[x].c = string(v)
		}
		fi[0].isEnter = true

		list = append(list, fi...)
		blockArr[blockIdx] = mixBlock(fi, tmpblock, blockArr[blockIdx])
	}


	rlt = make([]byte, 0, len(h))
	for _, v := range blockArr {

		rlt = append(rlt, v...)
	}

	return

}



func nextHFuncSite(buf []byte) (site int) {
	
	// var idx int
	bufSlice := buf
	// fmt.Println("nextFuncSite :" + string(cpp))
L:	for {

		idx := bytes.IndexFunc(bufSlice, indexEnd)
		if idx == -1 {
			return -1
		}

		if bufSlice[idx] == ';' {

			site = len(buf) - len(bufSlice) + idx + bytes.IndexByte(bufSlice[idx:], '\n') + 1
			break L
		}

		bufSlice = bufSlice[idx:]
		num := 0
		var line []byte
		for {
			
			if len(bufSlice) == 0 {
				return -1
			}

			idx = bytes.IndexByte(bufSlice, '\n')
			if idx == -1 {
				line = bufSlice
				bufSlice = []byte{}
			} else {
				line = bufSlice[:idx]
				bufSlice = bufSlice[idx+1:]
			}
			num = braceNum(line, num)

			if num == 0 {

				site = len(buf) - len(bufSlice)
				break L
			}
		}
	}
	return
}

func mixBlock(fi []*FuncInfo, tmpbuf []byte, buf []byte) (block []byte) {


	// nextHFuncSite()
	l, r := bytes.IndexByte(tmpbuf, '{'), bytes.LastIndexByte(tmpbuf, '}')
	l += bytes.IndexByte(tmpbuf[l:], '\n')+1
	r = bytes.LastIndexByte(tmpbuf[:r], '\n')+1

	block = make([]byte, l, len(buf))
	copy(block, buf)


	tmpSlice, bufSlice := tmpbuf[l:r], buf[l:r]
	oldFuncList := make([][]byte, 0, 10)
	curFuncList := fi
	N:	for {


		fiNum := len(curFuncList)
		if fiNum == 0 {
			pushLeftBuf(&block, oldFuncList)
			block = append(block, bufSlice...)
			break N
		}

		if len(oldFuncList) == 0 {
			
			site := nextHFuncSite(tmpSlice)
			if site == -1 {
	
				pushLeftRandHFunc(&block, curFuncList)
				pushLeftBuf(&block, oldFuncList)
				block = append(block, bufSlice...)
				break N
			}
			oldFuncList = append(oldFuncList, bufSlice[:site])
			tmpSlice, bufSlice = tmpSlice[site:], bufSlice[site:]
		}


		switch rand.Intn(2) {
			case 0:
				curFi := curFuncList[0]
				curFuncList = curFuncList[1:]
				block = append(block, []byte(curFi.ToNameInH())...)
			case 1:
				block = append(block, oldFuncList[0]...)
				oldFuncList = oldFuncList[1:]
		}
	}
	block = append(block, buf[r:]...)
	return
}

func pushLeftRandHFunc(rlt *[]byte, fl []*FuncInfo) {
	
	switch {
		case fl == nil: fallthrough
		case len(fl) == 0: return
	}

	for _, v := range fl {

		*rlt = append(*rlt, v.ToNameInH()...)
	}
}

func className(buf []byte) (cn []byte) {

	// istatic = bytes.Contains(buf, []byte("static "))

	idx := bytes.IndexByte(buf, '\n')
	line := buf[:idx]

	mao := bytes.IndexRune(line, ':')
	if mao == -1 {
		mao = bytes.IndexByte(line, '{')
		if mao == -1 {
			mao = len(line)
		}
	}

	class := []byte{'c', 'l', 'a', 's', 's', ' '}
	cn = bytes.TrimFunc(line[len(class):mao], trimLeftSpace)
	return
}


func nextClassSite(buf []byte) (lIdx int, rIdx int) {

	class := []byte{'c', 'l', 'a', 's', 's', ' '}
	bufSlice := buf
L:	for {

		r := bytes.IndexByte(bufSlice, '\n')
		if r == -1 {

			return -1, -1
		}

		line := bufSlice[:r]
		// fmt.Printf("line :%s\n", string(bufSlice))
		if !bytes.Contains(line, class) {
			bufSlice = bufSlice[r+1:]
			continue L
		}

		idx := bytes.IndexFunc(bufSlice, indexEnd)
		if idx == -1 || bufSlice[idx] == ';' {
			bufSlice = bufSlice[r+1:]
			continue L
		}

		lIdx = len(buf) - len(bufSlice)
		bufSlice = bufSlice[idx:]


		num := 0
		for {
			r = bytes.IndexByte(bufSlice, '\n')
			if r == -1 {
				return -1, -1
			}
			num = braceNum(bufSlice[:r], num)

			bufSlice = bufSlice[r+1:]
			if num == 0 {

				rIdx = len(buf) - len(bufSlice)
				break L
			}
		}
		
	}
	return
}

func indexEnd(r rune) (rlt bool) {

	rlt = false
	switch r {

		case ';': fallthrough
		case '{': rlt = true
	}
	
	return
}



// func mixHCode(list []*FuncInfo, code []byte) (rlt []byte) {

// 	codeLen := len(code)
// 	bio := bufio.NewReaderSize(bytes.NewReader(code), codeLen)
// 	rlt = make([]byte, 0, codeLen*2)
// 	listIdx := 0
// 	listLen := len(list)
// 	flag := 2 //0表示

// 	err := readNextFunc(bio)
// 	if err != nil {
// 		fmt.Println("处理类头文件时出错:" + err.Error())
// 		return
// 	}	
// 	rlt = append(rlt, code[:codeLen-bio.Buffered()]...)
	
// L:	for {

// 		var randNum int
// 		switch flag {
// 			case 0:
// 				randNum = 0
// 			case 1:					
// 				idx := codeLen - bio.Buffered()
// 				rlt = append(rlt, code[idx:]...)
// 				break L
// 			case 2: randNum = rand.Intn(2)
// 		}

// 		switch randNum {
// 			case 0:
// 				info := list[listIdx]
// 				str := info.ToNameInH()
// 				rlt = append(rlt, []byte(str)...)
// 				listIdx++
// 				if listIdx == listLen {
// 					if flag != 2 {
// 						break L
// 					}
// 					flag = 1
// 				}
// 			case 1:
// 				lastIdx := codeLen - bio.Buffered()
// 				err := readNextFunc(bio)
// 				switch err {
// 					case nil:
// 						curIdx := codeLen - bio.Buffered()
// 						rlt = append(rlt, code[lastIdx:curIdx]...)
// 					case io.EOF:
// 						rlt = append(rlt, code[lastIdx:]...)
// 						flag = 0
// 						continue L
// 					default:
// 						fmt.Println("there's an error when read line : " + err.Error())
// 						break L
// 				}
// 		}
// 	}
// 	return
// }

// // readNextFuncLine 用来读取下一个函数行
// func readNextFuncLine(bio *bufio.Reader) error {

// 	inNote, needCnt, defineCnt := false, 0, 0
// L:	for {
// 		line, _, e := bio.ReadLine()
// 		if e != nil {
// 			return e
// 		}

// 		usedLine := bytes.TrimLeftFunc(line, trimLeftSpace)
// 		if (len(usedLine) == 0) || !usefulLine(&usedLine, &inNote) || inNote {
// 			continue L
// 		}

// 		switch {
// 			case bytes.HasPrefix(usedLine, []byte("#if")): defineCnt++; continue L
// 			case bytes.HasPrefix(usedLine, []byte("#endif")): defineCnt--; continue L
// 			case defineCnt>0: continue L
// 			case bytes.HasPrefix(usedLine, []byte("#include")): continue L
// 		}
		
// 		if blockEnd(usedLine, &needCnt) {
// 			break L
// 		}
// 	}
// 	return nil
// }

// func splitHCode(d []byte) (l []byte, m []byte, r []byte, c []byte, istatic bool, nsBytes []byte) {

	
// 	ns, class, totalLen := []byte{'n', 'a', 'm', 'e', 's', 'p', 'a', 'c', 'e'}, []byte{'c', 'l', 'a', 's', 's'}, len(d)

// 	bio := bufio.NewReaderSize(bytes.NewReader(d), totalLen)
// 	lIdx, mIdx, innote := 0, 0, false
// 	needCnt, needTag := 0, false
// 	// nsBytes := nil

// 	L: 	for {

// 		leftLen := bio.Buffered()
// 		line, _, err := bio.ReadLine()
// 		switch err {
// 			case nil:
// 			case io.EOF:
// 				break L
// 			default:
// 				fmt.Println("there's an error when read line : " + err.Error())
// 				break L
// 		}
// 		usedLine := bytes.TrimFunc(line, trimLeftSpace)
// 		if (len(usedLine)==0) || !usefulLine(&usedLine, &innote) || innote {

// 			continue L
// 		}


// 		switch {
// 			case bytes.HasPrefix(usedLine, ns):
// 				tmp := usedLine[len(ns):]
// 				if tmp[len(tmp)-1] == '{' {
// 					tmp = tmp[:len(tmp)-1]
// 				}
// 				tmp = bytes.TrimFunc(tmp, trimLeftSpace)
// 				if nsBytes == nil {
// 					nsLen := len(tmp)
// 					nsBytes = make([]byte, nsLen, nsLen*2)				
// 				}
// 				nsBytes = append(nsBytes, tmp...)
// 				nsBytes = append(nsBytes, ':', ':')
// 				continue L
// 			case bytes.HasPrefix(usedLine, class) && !bytes.HasSuffix(usedLine, []byte{';'}):
// 				needCnt = braceNum(usedLine, needCnt)
// 				needTag = needCnt == 0
// 				if !needTag {
// 					lIdx = totalLen - bio.Buffered()
// 				}
// 				mao := bytes.IndexRune(usedLine, ':')
// 				if mao == -1 {
// 					if needTag {
// 						mao = len(usedLine)
// 					} else {
// 						mao = bytes.IndexRune(usedLine, '{')
// 					}
// 				}
// 				c = bytes.TrimFunc(usedLine[len(class):mao], trimLeftSpace)
// 				continue L
// 			case needTag: 
// 				needCnt = braceNum(usedLine, needCnt)
// 				if needCnt == 1 {
// 					needTag = false
// 					lIdx = totalLen - bio.Buffered()
// 				}
// 				continue L
// 			case lIdx == 0: continue L
// 		}

// 		needCnt = braceNum(usedLine, needCnt)
//         tmpIdx := bytes.IndexByte(usedLine, '(')
//         if tmpIdx != -1 {
//             usedLine = usedLine[:tmpIdx]
//         }
//         switch {
// 			case bytes.Contains(usedLine, []byte("static ")): istatic = true
// 			case needCnt == 0:
// 				mIdx = totalLen - leftLen
// 				break L
// 		}
// 	}

// 	if lIdx != 0 {	
// 		l, m, r = d[:lIdx], d[lIdx:mIdx], d[mIdx:]
// 	}
// 	return
// }
