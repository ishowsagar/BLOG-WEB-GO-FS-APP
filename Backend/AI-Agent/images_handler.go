package main

import (
	"encoding/base64"
	"os"
)

//& IMAGES READ BY AGENT
// *every thing is ultimately happening in bytes, from json to images...Since when we read image -> that comes as binary data
// * that binry data would crash it if not serialized properly or not encoded into safe text strings ( into 64 characters) casue
// * json data is text format with valid unicode characters but images's read binary data is just raw binary data, unsafe and would crash if sent as not decoded into desired variable or loaded into the struct
// ** that's when base64 emerges -> it takes that binaryData of image -> converts into safe text using bas64's encoding mechanism -> json compatible -> safe to put into any json field.

// cause when we send out encoded request -> it goes out as json data -> and then json is unmarshalad into structs as json data is safe -> won't crash
// but if we send raw image binary data -> and encode into json using marshalar it would crash -> as it expects text format which only be safely executed if this is base64 converted

// The internet is just bytes moving between machines all is just how data at bytes level would be as safely read and proccessed.

// func that reads image's binary into safe text string
func ReadIMGBinary(imgName string)([]byte,error) {

	// os.ReadFile -> reads file directly under the hood by opening it and then read into buf -> accumulated byte
	image,err := os.ReadFile(imgName)
	if err != nil {
		return nil,err
	}

	// returning image []of bytes data 
	return image,nil 
	
}


// func that converts image binary bytes data into safe printable systems recognized string formatted
func ImageBinaryConversion(binary []byte) (safeIMGTextFormatted string) { // return named are intialized var  -> must be assigned for return

	//ohhhh base64 is for conversion of bytes data into printable text format -> now its high time to get feel of this topic too
	safeIMGTextFormatted = base64.StdEncoding.EncodeToString(binary)
	return safeIMGTextFormatted
}


// this func takes in base64 encoded safe str and converts back into raw binary which then have to be populated into img file -> for displaying image actually
func ImageEncoder(base64SafeImgStr string) (decodedBinaryBytes[]byte,err error) {
	
	decodedBinaryBytes,err =base64.StdEncoding.DecodeString(base64SafeImgStr)
	if err != nil {
		return nil,err
	}

	// if successfully encoded back to binary bytes data ranges from 0-255
	return decodedBinaryBytes,nil
}