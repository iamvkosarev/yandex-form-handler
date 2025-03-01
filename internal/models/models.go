package models

type FormResult struct {
	CouchResult, ClientResult ResultEmailData
}

type ResultEmailData struct {
	Destination, BodyText, BodyHTML string
}
