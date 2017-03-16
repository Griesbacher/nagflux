package crypto

import (
	"testing"
)

var plain = `DATATYPE::SERVICEPERFDATA	TIMET::1489564463	HOSTNAME::dummyHost	SERVICEDESC::load	SERVICEPERFDATA::load1=0.090;1.000;2.000;0; load5=0.100;5.000;10.000;0; load15=0.090;15.000;30.000;0;	SERVICECHECKCOMMAND::check_local_load!1,5,15!2,10,30	SERVICESTATE::0	SERVICESTATETYPE::1
SERVICEINTERVAL::1.000000

` + string([]rune{'\x00', '\x00', '\x00', '\x00'})

const cypher = `W6brRuzUSGFMjsddHulCbHRaHLCMYD40YD67LKD/zzFyqvonQilrtPkStkdLc3gtA675Il3QAK2BJnGCA6iP05y+9OLXGEOIfibCh8sOITacCOkF0XfyBv2qEQmjkdA8iSiqO5hFxPqyZbMIhzFJU1cQ1EszAAT+2vuG/IjqXSY9i9l6a/I3p/M6uQB/mFDhwqnV6NmfeRyQ0REKTCuv3ywnzwPci/90GpI6Vwn5bBNlVk8pi6cYcjJG7JaZ8oMWn3M6Q+zP5zfA+6lYKItwTmy7hf/ekGPV7dxkUaFSm5HMc2BKXZdfLYxfp8LIuH+gutIEJjEtsxY99kwq20/hUyiDkAg5gNf2mSQUNCfEwcpBwy5UMKoBJOG6es7VFB1T+PrPFdPdtxhr7zOS9Ws+GA==`
const key = `ac4tgMnAZhhUytwdTMJHnEtTbFMrVja`

func TestNewAESECBDecrypter(t *testing.T) {
	t.Parallel()
	pt, err := NewAESECBDecrypter([]byte("abcdefghijklmnopqrstuvwxyz"))
	if pt != nil && err == nil {
		t.Error("This key should not be valid")
	}
	pt, err = NewAESECBDecrypter([]byte("abcdefghijklmnopqrstuvwx"))
	if pt == nil && err != nil {
		t.Error("This key should be valid: err:", err)
	}
}

func TestAESECBDecrypter_Decypt(t *testing.T) {
	t.Parallel()
	pt, err := NewAESECBDecrypter([]byte(key + string([]rune{'\x00'})))
	if pt == nil && err != nil {
		t.Error("This key should be valid: err:", err)
	}
	result, err := pt.Decypt([]byte(cypher))
	if err != nil {
		t.Error(err)
	}
	if string(result) != plain {
		t.Error("The decrypted did not match the crypted")
	}
	result, err = pt.Decypt([]byte("123"))
	if result != nil && err == nil {
		t.Error("There should be no result: result:", result)
	}
}
