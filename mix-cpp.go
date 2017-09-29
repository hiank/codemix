package codemix

import (
	"fmt"
	"time"
	"strconv"
	"strings"
	"math/rand"
	"bytes"
)


func mixCpp(fl []*FuncInfo, cpp []byte) (rlt []byte) {

	rand.Seed(time.Now().UnixNano())

	endIdx := bytes.LastIndexByte(cpp, '}')
	if endIdx == -1 {
		return cpp
	}

	fm := toFuncInfoMap(fl)

	var ct []byte

	tmp := cleanBuf(cpp)

    tmpIdx := bytes.IndexByte(tmp, '}')
    tmpIdx = bytes.IndexByte(tmp[endIdx:], '\n')
    var endCpp []byte
    if tmpIdx != -1 {
        endIdx += tmpIdx + 1
        endCpp = cpp[endIdx:]
        cpp = cpp[:endIdx]
        tmp = tmp[:endIdx]
    }


	tmpSlice, cppSlice := tmp, cpp
	oldFuncList := make([][]byte, 0, 10)
	var curFuncList []*FuncInfo
	N:	for {

		l, r := nextFuncSite(tmpSlice)
		if l == -1 {
			pushLeftRandFunc(&rlt, curFuncList)
			pushLeftBuf(&rlt, oldFuncList)
			rlt = append(rlt, cppSlice...)
			break N
		}
		nfT, nfCpp := tmpSlice[l:r], cppSlice[l:r]
		rlt = append(rlt, cppSlice[:l]...)
		tmpSlice, cppSlice = tmpSlice[r:], cppSlice[r:]

		curCt, tag := classTag(nfT), 0
		switch {
			case bytes.Compare(curCt, ct) != 0:
				if ct != nil {
					fm[string(ct)] = nil
				}
				ct = curCt
				pushLeftRandFunc(&rlt, curFuncList)
				// fm[string(ct)] = nil
				pushLeftBuf(&rlt, oldFuncList)
				oldFuncList = make([][]byte, 0, 10)
				curFuncList = fm[string(curCt)]
				if curFuncList != nil && len(curFuncList) > 0 {
					
					call := toCallLine(curFuncList[0], curCt, nfT)
					nf := insertFuncCall(nfT, nfCpp, call)
					rlt = append(rlt, nf...)
					continue N
				}
				fmt.Printf("no class name %s___%v\n", string(curCt), fm)
				tag = 1
			case curFuncList == nil: fallthrough
			case len(curFuncList) == 0: tag = 1
			default: tag = rand.Intn(2)
		}

		oldFuncList = append(oldFuncList, nfCpp)
		switch tag {
			case 0:
				curFi := curFuncList[0]
				var con string
				if curFi.isEnter {
					con = toCallType(curFuncList)
				} else {
					con = toBaseType(curFi)
				}
				curFuncList = curFuncList[1:]
				fm[string(ct)] = curFuncList
				rlt = append(rlt, []byte(con)...)
			case 1:
				rlt = append(rlt, oldFuncList[0]...)
				oldFuncList = oldFuncList[1:]
		}
	}

    if endCpp != nil {
        rlt = append(rlt, endCpp...)
    }
	return
}

func toCallLine(info *FuncInfo, cn []byte, nf []byte) []byte {

	head := bytes.TrimFunc(nf[:bytes.IndexByte(nf, '{')], trimFiled)

	switch {
        case bytes.HasSuffix(head, []byte(" const")): return []byte{}

	}

	pre := "\n\n\n\tint _mix_a = 7;\n\tint _mix_b = 4;\n\tbool _mix_sd = false;\n\tfor(int c=0; c<_mix_b*0.5; c++) {\n\t\t_mix_b += 1;\n\t}\n\t_mix_sd = _mix_a > _mix_b;\n\tif (_mix_sd)\n"
	if info.istatic {
		pre += "\t\t" + info.c + "::"
	} else {
		pre += "\t\tthis->"
	}
	return []byte(pre + info.n + "(" + paramValue(&info.p) + ");\n")
}


func insertFuncCall(fnT []byte, fnCpp []byte, call []byte) []byte {

	// call := []byte("\tthis." + info.n + "(" + paramValue(&info.p) + ");\n")
	callLen := len(call)
	maxLen := len(fnT)+callLen
	nfn := make([]byte, maxLen)

	idx := bytes.LastIndex(fnT, []byte("return "))
	if idx != -1 {
		idx = bytes.LastIndexByte(fnT[:idx], '\n') + 1
	} else {
		idx = bytes.LastIndexByte(fnT, '}')
	}

	copy(nfn, fnCpp[:idx])
	copy(nfn[idx:], call)
	copy(nfn[idx+callLen:], fnCpp[idx:])

	return nfn
}


func pushLeftRandFunc(rlt *[]byte, fl []*FuncInfo) {

	switch {
		case fl == nil: fallthrough
		case len(fl) == 0: return
	}

	if fl[0].isEnter {

		*rlt = append(*rlt, toCallType(fl)...)
		fl = fl[1:]
	}
	for _, v := range fl {

		*rlt = append(*rlt, toBaseType(v)...)
	}
}


func classTag(nf []byte) []byte {

	// fmt.Println("classTag :" + string(nf))
	r := bytes.LastIndex(nf[:bytes.IndexByte(nf, '(')], []byte{':', ':'})
	nf = nf[:r]
	l := bytes.LastIndexFunc(nf, func(re rune) bool {
		rlt := false
		switch re {
			case '*': rlt = true
			default: rlt = trimFiled(re)
		}
		return rlt
	}) + 1

	return nf[l:]
}


func nextBraceIndex(buf []byte) (idx int) {

	num := 0
L:	for {
		
		bufLen, r := len(buf), bytes.IndexByte(buf, '\n')
		if r == -1 {
			r = bufLen - 1
		}
		// fmt.Printf("r___:%d\n", r)
		num = braceNum(buf[:r], num)
		r++
		idx += r
		// fmt.Printf("r:%d__bufLen:%d\n", r, bufLen)
		switch {
			case num == 0: fallthrough
			case r == bufLen: break L
		}
		buf = buf[r:]
	}
	return
}


func toFuncInfoMap(fiList []*FuncInfo) map[string][]*FuncInfo {

	// var list []string
	m := make(map[string][]*FuncInfo)
	maxLen := len(fiList)
	for _, v := range fiList {

		list := m[v.c]
		if list == nil {
			list = make([]*FuncInfo, 0, maxLen)
			v.isEnter = true
		}
		m[v.c] = append(list, v)
	}

	return m

}


func toCallType(fiList []*FuncInfo) (content string) {

	info := fiList[0]
	content = "\n" + info.t + " " + info.c + "::" + info.n + "(" + info.p + ") {\n"
	var pre string
	if info.istatic {
		pre = "\t" + info.c + "::"
	} else {
		pre = "\tthis->"
	}
	for _, v := range fiList[1:] {

		content += pre + v.n + "(" + paramValue(&v.p) + ");\n"
	}
	content += "}\n"
	return
}

func paramValue(p *string) (pv string) {

	pt, pn := paramList(p)
	if pt == nil {
		return
	}
	
	for i, v := range pt {

		var t string
		switch v {
			case "const char*": t = "\"" + pn[i] + "\""
			case "int": t = strconv.Itoa(rand.Intn(1024))
			case "float": t = strconv.FormatFloat(rand.Float64()*666, 'f', 2, 64)
		}
		pv += ", " + t
	}
	pv = pv[2:]
	return
}

// 将字符串分解成两个数组，pt为类型数组，pn为变量名数组
func paramList(p *string) (pt []string, pn []string) {

	if *p == "" {
		return
	}

	arr := strings.Split(*p, ", ")
	pt, pn = make([]string, len(arr)), make([]string, len(arr))
	for i, v := range arr {

		last := strings.LastIndexByte(v, ' ')
		pt[i], pn[i] = v[:last], v[last+1:]
	}
	return
}

func toBaseType(info *FuncInfo) (content string) {

	if info.c == "" {
		return ""
	}
	content = "\n" + info.t + " " + info.c + "::" + info.n + "(" + info.p + ") {\n"

	pt, pn := paramList(&info.p)

	tag := 0
	if pt != nil {
		tag = len(pt)
	}
	switch info.t {
		case "void":
			str := "\n\tCCLOG(\""
			switch tag {
				case 0: str += info.n + "\""
				default:
					var n string
					for i, v := range pt {
						var t string
						switch v {
							case "const char*": t = "%s"
							case "int": t = "%d"
							case "float": t = "%.2f"
						}
						str += t
						n += ", " + pn[i]
					}
					str += "\"" + n
			}
			content += str + ");\n"
		case "int": fallthrough
		case "float":
			var r string
			switch tag {
				case 0: r = "0"
				default:
					tmp := make([]string, 0, tag) 
					for i, v := range pt {
						switch v {
							case "int": fallthrough
							case "float": tmp = append(tmp, pn[i])
						}
					}
					switch len(tmp) {
						case 0: r = "0"
						case 1: r = tmp[0]
						case 2: r = tmp[0] + " * " + tmp[1]
						default:
							r = tmp[0]
							tmp = tmp[1:]
							for _, v := range tmp {

								r += " + " + v
							}
					}
			}
			content += "\n\treturn " + r + ";\n"
		case "const char*":
			var r string
			switch tag {
				case 0: content += "\n\treturn \"" + info.n + "\";\n" 
				default:
					r = "\n\tchar str[512];\n\tsprintf(str, \""
					var n string
					for i, v := range pt {
						var t string
						switch v {

							case "const char*": t = "%s"
							case "int": t = "%d"
							case "float": t = "%.2f"
						}
						r += t
						n += ", " + pn[i]
					}
					r += "\"" + n + ");\n"
					content += r + "\treturn str;\n"
			}
	}
	content += "}\n"
	return
}


func nextFuncSite(cpp []byte) (l int, r int) {
	
	var idx int
	cppSlice := cpp
	// fmt.Println("nextFuncSite :" + string(cpp))
	show := false
	for {

		idx = bytes.IndexByte(cppSlice, '{')
		if idx == -1 {
			return -1, -1
		}

		lCpp := bytes.TrimRightFunc(cppSlice[:idx], trimFiled)
		lIdx := bytes.LastIndexByte(lCpp, '\n') + 1
		name := lCpp[lIdx:]
		nIdx := bytes.IndexByte(name, '(')
		// fmt.Println("name is :" + string(name))
		if nIdx!=-1 && bytes.Contains(name[:nIdx], []byte{':', ':'}) {
			if lIdx > 0 {
				tmp := lCpp[:lIdx-1]
				tmpIdx := bytes.LastIndexByte(tmp, '\n') + 1
				if tmpIdx != 0 {

					tmp = tmp[tmpIdx:]
					if bytes.HasPrefix(tmp, []byte("template")) {
						show = true
						lIdx = tmpIdx
					}
				}
			}
			l = len(cpp) - len(cppSlice) + lIdx
			break
		}
		cppSlice = cppSlice[nextBraceIndex(cppSlice[idx:]):]
	}

	r = len(cpp) - len(cppSlice) + nextBraceIndex(cppSlice[idx:]) + idx
	if show {

		fmt.Println("fun is :" + string(cpp[l:r]))
	}
	return
}
