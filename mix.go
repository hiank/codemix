package codemix

import (
	"strings"
	"github.com/hiank/codemix/config"
	"bufio"
	"fmt"
	"bytes"
	"math/rand"
)


// FuncInfo is the struct of function info in class
type FuncInfo struct {
	c string

	t string
	n string
	p string

	isEnter bool
	istatic bool
}

// ToNameInH 获取函数在h 文件中的全名
func (info *FuncInfo)ToNameInH() string {

    var pre string
    if info.istatic {
        pre = "\tstatic "
    } else {
        pre = "\t"
    }
	return pre + info.t + " " + info.n + "(" + info.p + ");\n"
}

// // ToCPP 获取函数在cpp 文件中的内容
// func (info *FuncInfo)ToCPP() (content string) {

// 	if info.c == "" {
// 		return ""
// 	}
// 	content = "\n" + info.t + " " + info.c + "::" + info.n + "(" + info.p + ") {\n"

// 	var arr, pArr []string
// 	if info.p != "" {
		
// 		arr = strings.Split(info.p, ", ")
// 		pArr = make([]string, len(arr))
// 		for i, v := range arr {
// 			pArr[i] = v[strings.LastIndex(v, " ")+1:]
// 		}
// 	}


// 	tag := 0
// 	if pArr != nil {
// 		tag = len(pArr)
// 	}
// 	switch info.t {
// 		case "void":
// 			str := "\n\tCCLOG(\""
// 			switch tag {
// 				case 0: str += info.n + "\""
// 				default:
// 					var n string
// 					for i, v := range arr {
// 						var t string
// 						switch v[:strings.LastIndex(v, " ")] {

// 							case "const char*": t = "%s"
// 							case "int": t = "%d"
// 							case "float": t = "%.2f"
// 						}
// 						str += t
// 						n += ", " + pArr[i]
// 					}
// 					str += "\"" + n
// 			}
// 			content += str + ");\n"
// 		case "int": fallthrough
// 		case "float":
// 			var r string
// 			switch tag {
// 				case 0: r = "0"
// 				// case 1: r = pArr[0]
// 				// case 2: r = pArr[0] + " * " + pArr[1]
// 				default:
// 					tmp := make([]string, 0, tag) 
// 					for i, v := range arr {
// 						switch v[:strings.LastIndex(v, " ")] {
// 							case "int": fallthrough
// 							case "float": tmp = append(tmp, pArr[i])
// 						}
// 					}
// 					switch len(tmp) {
// 						case 0: r = "0"
// 						case 1: r = tmp[0]
// 						case 2: r = tmp[0] + " * " + tmp[1]
// 						default:
// 							r = tmp[0]
// 							tmp = tmp[1:]
// 							for _, v := range tmp {

// 								r += " + " + v
// 							}
// 					}
// 			}
// 			content += "\n\treturn " + r + ";\n"
// 		case "const char*":
// 			var r string
// 			switch tag {
// 				case 0: content += "\n\treturn \"" + info.n + "\";\n" 
// 				default:
// 					r = "\n\tchar str[512];\n\tsprintf(str, \""
// 					var n string
// 					for i, v := range arr {
// 						var t string
// 						switch v[:strings.LastIndex(v, " ")] {

// 							case "const char*": t = "%s"
// 							case "int": t = "%d"
// 							case "float": t = "%.2f"
// 						}
// 						r += t

// 						n += ", " + pArr[i]
// 					}
// 					r += "\"" + n + ");\n"
// 					content += r + "\treturn str;\n"
// 			}
// 	}
// 	content += "}\n"
// 	return
// }

// Mix used to mix the h and cpp file data
func Mix(h []byte, cpp []byte) ([]byte, []byte) {
	
	// rand.Seed(time.Now().UnixNano())
	// num := 20 + rand.Intn(80)

	
	// randFunc(funcList)
	h, funcList := mixH(h)

	// fmt.Printf("mixH end funcList :%v\n", funcList)
	cpp = mixCpp(funcList, cpp)

	return h, cpp
}

func randFunc(list []*FuncInfo) {

	sameName := func(l []*FuncInfo, name string) (rlt bool) {

		rlt = false
		for _, v := range l {

			if v.n == name {
				rlt = true
				break				
			}
		}
		return
	}

	for i := range list {

		t, n, p := randType(4), randName(), randParam()
		tmp := list[:i]
		for sameName(tmp, n) {

			n = randName()
		}
		f := new(FuncInfo)
		*f = FuncInfo{
			t: t,
			n: n,
			p: p,
		}
		list[i] = f
	}	
}

func randType(num int) (t string) {

	switch rand.Intn(num) {
		case 0: t = "const char*"
		case 1: t = "int"
		case 2: t = "float"
		case 3: t = "void"
		default: t = "void"
	}
	return
}

func randName() (n string) {


	ignore := config.GetConfig().M.IgnoreRandname

L:	for {
		nameLen := 10 + rand.Intn(8)
		t := make([]byte, nameLen)
		for i := range t {

			t[i] = randEnByte()
		}
		n = string(t)
		
		for _, v := range ignore {
			
			if strings.Contains(n, v) {
	
				continue L
			}
		}
		break L
	}
	return
}

func randNumByte() byte {

	return byte(int('0') + rand.Intn(10))
}

func randEnByte() byte {

	idx := rand.Intn(52)
	var base int
	if idx < 26 {
		base = 65
	} else {
		base = 97 - 26
	}
	return byte(base + idx)
}

func randParam() (p string) {

	num := rand.Intn(4)
	nameList := make([]string, 0, num)
	if num > 0 {
		for i:=0; i<num; i++ {

			// p += randType() + " p" + strconv.Itoa(i) + ", "
			name := randParamName(nameList)
			p += randType(3) + " " + name + ", "
			nameList = append(nameList, name)
		}
		p = p[:len(p)-2]
	}
	return
}

func randParamName(list []string) (name string) {

	ignore := config.GetConfig().M.IgnoreRandname

L:	for {
		nameLen := 2 + rand.Intn(6)
		numLen := 1 + rand.Intn(3)
		t := make([]byte, nameLen, nameLen + numLen)
		for i := range t {

			t[i] = randEnByte()
		}
		for i:=0; i<numLen; i++ {
			
			t = append(t, randNumByte())
		}
		
		name = string(t)
		for _, v := range ignore {
			
			if strings.Contains(name, v) {
	
				continue L
			}
		}
			
		for _, v := range list {

			if v == name {
				
				continue L
			}
		}
		break L
	}
	return
}



// braceNum 计算需要的大括号右括号数量
func braceNum(line []byte, oldNum int) (rlt int) {

	idx := bytes.Index(line, []byte{'/', '/'})
	if idx != -1 {
		line = line[:idx]
	}
	lNum := bytes.Count(line, []byte{'{'})
	rNum := bytes.Count(line, []byte{'}'})
	
	rlt = oldNum + lNum - rNum
	return
}

func trimLeftSpace(r rune) (rlt bool) {

	switch r {
		case 0: fallthrough
		case ' ': fallthrough
		case '\t': rlt = true
		default: rlt = false
	}
	return
}

func usefulLine(line *[]byte, status *bool) (useful bool) {

	r := bytes.Index(*line, []byte{'*', '/'})
	if *status {
		if r != -1 {
			*status = false
		}
		return
	}

	l, one := bytes.Index(*line, []byte{'/', '*'}), bytes.Index(*line, []byte{'/', '/'})
	switch {
		case r != -1:
			r += 2
			lineLen := len(*line)
			if l != -1 && l < r && r < lineLen {
				tmp := make([]byte, lineLen-r+l)
				copy(tmp, (*line)[:l])
				copy(tmp[l:], (*line)[r:])
				*line = bytes.TrimFunc(tmp, trimLeftSpace)
				useful = true
			}
		case one != -1:
			if one != 0 {
				*line = bytes.TrimFunc((*line)[:one], trimLeftSpace)
				useful = true
			}
		case l != -1: *status = true
		default: useful = true
	}
	return
}
// // InNote 用于清除io中最开始的注释段
// type InNote func(*io.Reader)

// blockEnd 用于判断line 行是否是闭合段，cnt用于保存左括号的数量
func blockEnd(line []byte, cnt *int) (end bool) {

	curCnt := braceNum(line, *cnt)
	switch {
		case *cnt > 0:
			*cnt = curCnt
			if curCnt == 0 {
				end = true
			}
		case *cnt < 0:
			fmt.Printf("blockEnd error, 左括号数量异常: %d, %s\n", *cnt, string(line))
			end = false
		case *cnt == 0:
			preKeyword := [][]byte{
				[]byte{'e', 'n', 'u', 'm'}, 
				[]byte{'u', 'n', 'i', 'o', 'n'},
			}
			sufKeyword := [][]byte{
				[]byte{'N', 'S', '_', 'C', 'C', '_', 'B', 'E', 'G', 'I', 'N'},
				[]byte{'N', 'S', '_', 'C', 'C', '_', 'E', 'X', 'T', '_', 'B', 'E', 'G', 'I', 'N'},
				[]byte{'N', 'S', '_', 'C', 'C', '_', 'E', 'N', 'D'},
				[]byte{'N', 'S', '_', 'C', 'C', '_', 'E', 'X', 'T', '_', 'E', 'N', 'D'},
				[]byte{' ', 'c', 'o', 'n', 's', 't'},
				[]byte{')'},
				[]byte{'>'},
			}
			switch {
				case curCnt > 0: *cnt = curCnt; fallthrough
				case _prefix(line, preKeyword): fallthrough
				case _suffix(line, sufKeyword): end = false
				default: end = true
			}
	}
	return
}

func _prefix(line []byte, arr [][]byte) (rlt bool) {
	
	for _, v := range arr {
		
		if bytes.HasPrefix(line, v) {
			rlt = true
			break
		}
	}
	return
}

func _suffix(line []byte, arr [][]byte) (rlt bool) {

	for _, v := range arr {

		if bytes.HasSuffix(line, v) {
			rlt = true
			break
		}
	}
	return
}

func readNextFunc(bio *bufio.Reader) error {

	inNote, defineCnt, needCnt := false, 0, 0
L:	for {
		line, _, e := bio.ReadLine()
		if e != nil {
			return e
		}
		usedLine := bytes.TrimFunc(line, trimLeftSpace)
		if (len(usedLine)==0) || !usefulLine(&usedLine, &inNote) || inNote {
			continue L
		}
		switch {
			case bytes.HasPrefix(usedLine, []byte("#if")): defineCnt++; continue L
			case bytes.HasPrefix(usedLine, []byte("#endif")): defineCnt--; continue L
			case defineCnt>0: continue L
			case bytes.HasPrefix(usedLine, []byte("#include")): continue L
		}
		if blockEnd(usedLine, &needCnt) {
			break L
		}

	}
	return nil
}


// func nextActiveField(buf []byte) (l int, r int) {

// 	bufLen, tmp := len(buf), bytes.TrimLeftFunc(buf, trimFiled)
// 	baseIdx := bufLen - len(tmp)
// 	x, y := bytes.Index(tmp, []byte{'/', '/'}), bytes.Index(tmp, []byte{'/', '*'})
// 	switch {
// 		case x == y: l, r = 0, bufLen-1
// 		case y == 0:
// 			idx := bytes.Index(tmp, []byte{'*', '/'}) + 2
// 			l, r = nextActiveField(tmp[idx:])
// 		case x == 0:
// 			l, r = nextActiveField(tmp[bytes.IndexByte(tmp, '\n')+1:])
// 		default:
// 			if x < y {
// 				r, l = x, y
// 			} else {
// 				r, l = y, x
// 			}
// 			if r == -1 {
// 				r = l
// 			}
// 			l = 0
// 	}
// 	l, r = l + baseIdx, r + baseIdx
// 	return
// }

// 获取最前一段注释位置
func nextNoteField(buf []byte) (l int, r int) {

	tagFunc := func(ru rune) (rlt bool) {

		switch ru {

			case '/': fallthrough
			case '#': rlt = true
		}
		return
	}

	// var lIdx int
	rBuf, bufLen := buf, len(buf)
L:	for {

		rBuf = bytes.TrimLeftFunc(rBuf, trimFiled)

		lIdx := bytes.IndexFunc(rBuf, tagFunc)
		if lIdx == -1 {
			return -1, -1
		}
		rBuf = rBuf[lIdx:]
		switch rBuf[0] {
			case '/':
				switch rBuf[1] {
					case '*':
						l = bufLen - len(rBuf)
						r = l + bytes.Index(rBuf, []byte{'*', '/'}) + 2
						break L
					case '/':
						goto TOEND
					default: 
						rBuf = rBuf[1:]
						continue L
				}
			case '#':
				switch {
					// case bytes.HasPrefix(rBuf[1:], []byte{'i', 'f', 'd', 'e', 'f'}):
					// 	l = bufLen - len(rBuf)
					// 	rBuf = rBuf[6:]

					case bytes.HasPrefix(rBuf, []byte("#ifdef")): fallthrough
					case bytes.HasPrefix(rBuf[1:], []byte{'i', 'f', ' '}):
						l = bufLen - len(rBuf)
						rBuf = rBuf[3:]
						needCnt := 1
						ENDTAG: for {

							eIdx := bytes.Index(rBuf, []byte{'#', 'e', 'n', 'd', 'i', 'f'})
							bIdx := bytes.Index(rBuf, []byte("#if"))

							if bIdx != -1 && eIdx > bIdx {

								needCnt++
								rBuf = rBuf[bIdx+3:]
								continue ENDTAG
							}
							needCnt--
							rBuf = rBuf[eIdx+6:]
							if needCnt == 0 {

								break ENDTAG
							}
						}
						r = bufLen - len(rBuf)
						break L
					default:
						goto TOEND
				}
		}
	}
	return

TOEND:
	tmp := bytes.IndexByte(rBuf, '\n')
	switch {
		case tmp == -1: tmp = len(rBuf)
		case rBuf[tmp-1] == '\r': tmp--
	}
	l = bufLen - len(rBuf)
	r = l + tmp
	return
}


func trimFiled(r rune) (rlt bool) {

	switch r {
		case 0: fallthrough
		case ' ': fallthrough
		case '\r': fallthrough
		case '\n': fallthrough
		case '\t': rlt = true
	}
	return
}


func cleanBuf(buf []byte) []byte {

	tmp := make([]byte, len(buf))
	tmpSlice, bufSlice := tmp, buf
	L:	for {

		if len(tmpSlice) == 0 {
			break L
		}

		fl, fr := nextNoteField(bufSlice)
		if fl == -1 {
			
			copy(tmpSlice, bufSlice)
			break L
		}

		if fl != 0 {
			
			copy(tmpSlice[:fl], bufSlice[:fl])
		}
		tmpSlice, bufSlice = tmpSlice[fr:], bufSlice[fr:]
	}
	return tmp
}


func pushLeftBuf(rlt *[]byte, list [][]byte) {
	
	switch {
		case list == nil: fallthrough
		case len(list) == 0: return
	}

	for _, v := range list {

		*rlt = append(*rlt, v...)
	}
}
	
