package rego

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"io"
	"log"
	"net"
)

//webSocket相关
type WS struct{
	wsMap map[string]WSConfig
	WSCon
}

type WSConfig struct{
	OnOpen func(con *WSCon)
	OnClose func(con *WSCon)
	OnMessage func(con *WSCon,data []byte)
	OnError func(con *WSCon)
}

type WSCon struct{
	Conn net.Conn
	MaskingKey []byte
	Buf  *bufio.ReadWriter
}

func (this *WS)WebSocket(path string,config WSConfig){
	this.wsMap[path]=config
}

func  (this *WSCon)handleConnection(c *Context,wsConfig WSConfig){
	var err error
	method:=c.Request.Method
	if method!="GET"{
		return
	}else{
		request:=c.Request
		secWebSocketKey:=request.Header.Get("Sec-WebSocket-Key")
		guid:="258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
		h:=sha1.New()
		_, err=io.WriteString(h,secWebSocketKey+guid)
		accept:=make([]byte,28)
		base64.StdEncoding.Encode(accept,h.Sum(nil))
		response := "HTTP/1.1 101 Switching Protocols\r\n" +
			"Upgrade: websocket\r\n" +
			"Connection: Upgrade\r\n" +
			"Sec-WebSocket-Accept: " + string(accept) + "\r\n\r\n"
		if _,err:=this.Buf.Write([]byte(response));err!=nil{
			CheckErr(err,"write into the connection failed")
		}
		err = this.Buf.Flush()
		if err != nil {
			CheckErr(err,"the buf flush failed")
		}
		wsConfig.OnOpen(this)
		go func() {
			for{
				data,err:=this.ReadIframe(wsConfig)
				if err != nil {
					CheckErr(err,"the err on the readIframe ")
					wsConfig.WSClose(this)
					break
				}
				log.Println("read data:",string(data))
				wsConfig.OnMessage(this,data)
				err=this.SendIframe([]byte("good"))
				if err != nil {
					CheckErr(err,"the err on the sendIframe")
					wsConfig.WSClose(this)
					break
				}
				log.Println("send data")
			}
		}()
	}
}

func (this *WSCon)ReadIframe(wsConfig WSConfig)(data []byte,err error){
	opcodeByte := make([]byte, 1)
	_, err =this.Buf.Read(opcodeByte)

	FIN := opcodeByte[0] >> 7
	RSV1 := opcodeByte[0] >> 6 & 1
		RSV2 := opcodeByte[0] >> 5 & 1
		RSV3 := opcodeByte[0] >> 4 & 1
		if RSV1==1 || RSV2==1 || RSV3==1{
			wsConfig.WSClose(this)
		return
	}

	payloadLenBt := make([]byte, 1)
	_, err =this.Buf.Read(payloadLenBt)
	payloadLen := int(payloadLenBt[0] & 0x7F)
	mask := payloadLenBt[0] >> 7
	if payloadLen == 127 {
		extendedByte := make([]byte, 8)
		_, err=this.Buf.Read(extendedByte)
	}

	maskingByte := make([]byte, 4)
	if mask == 1 {
		_, err=this.Buf.Read(maskingByte)
		this.MaskingKey = maskingByte
	}

	payloadDataByte := make([]byte, payloadLen)
	_, err=this.Buf.Read(payloadDataByte)

	dataByte := make([]byte, payloadLen)
	for i := 0; i < payloadLen; i++ {
		if mask == 1 {
			dataByte[i] = payloadDataByte[i] ^ maskingByte[i % 4]
		} else {
			dataByte[i] = payloadDataByte[i]
		}
	}

	if FIN == 1 {
		data = dataByte
		return
	}

	nextData, err := this.ReadIframe(wsConfig)
	if err != nil {
		return
	}
	data = append(data, nextData...)

	return
}

func (this *WSCon)SendIframe(data []byte)(err error){
	lenth := len(data)
	maskedData := make([]byte, lenth)
	for i := 0; i < lenth; i++ {
		if this.MaskingKey != nil {
			maskedData[i] = data[i] ^ this.MaskingKey[i % 4]
		} else {
			maskedData[i] = data[i]
		}
	}
	_, err =this.Buf.Write([]byte{0x81})
	var payLenByte byte
	if this.MaskingKey != nil && len(this.MaskingKey) != 4 {
		payLenByte = byte(0x80) | byte(lenth)
		_, err =this.Buf.Write([]byte{payLenByte})
		_, err =this.Buf.Write(this.MaskingKey)
	} else {
		payLenByte = byte(0x00) | byte(lenth)
		_, err =this.Buf.Write([]byte{payLenByte})
	}
	_, err =this.Buf.Write(data)
	err =this.Buf.Flush()
	return
}


func (this *WSConfig)WSClose(wsCon *WSCon){
	this.OnClose(wsCon)
	_ = wsCon.Conn.Close()
}

