package client

import (
	//"errors"
	"github.com/serdmanczyk/freyr/models"
	//"github.com/serdmanczyk/freyr/oauth"
	//"github.com/serdmanczyk/freyr/token"
	"net/http"
	//"time"
)

type FreyrApiClient struct {
	client *http.Client
	Domain string
	Secret models.Secret
}

//|     route      | web | api | device |
//|----------------|-----|-----|--------|
//| /              | x   | x   |        |
//| /secret        | x   | x   |        |
//| /latest        | x   | x   |        |
//| GET /readings  | x   | x   |        |
//| POST /reading  |     | x   | x      |
//| /rotate_secret |     | x   |        |
