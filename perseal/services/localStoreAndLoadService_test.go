package services

import (
	"log"
	"testing"

	"github.com/EC-SEAL/perseal/dto"
	"github.com/EC-SEAL/perseal/sm"
	"github.com/EC-SEAL/perseal/utils"
)

func TestStoreService(t *testing.T) {

	obj := InitIntegration("Browser")

	session, _ := sm.GetSessionData(obj.ID)
	obj, _ = dto.PersistenceWithPasswordBuilder(obj.ID, session, "qwerty")
	log.Println(session)
	ds, err := PersistenceStore(obj)
	log.Println(ds)
	log.Println(err)
	if err != nil {
		t.Error("Thrown error, got: ", err)
	}

	session, _ = sm.GetSessionData(obj.ID)
	sha := utils.HashSUM256("qwerty")
	obj, _ = dto.PersistenceWithPasswordBuilder(obj.ID, session, sha)
	log.Println(session)
	obj.LocalFileBytes = []byte(`"{\"id\":\"DS_521f98d6-7578-41bd-bbe3-7a3eeb35e1e1\",\"encryptedData\":\"a5zwnD-D1A1R3ZGlprOLqwKO7L_Ww0CRgqzqa4tkczRGqoCAPFmdZepv56HWDsGknfEUOXH_X0Am4tPQhurjS-X2Qq0yARP-Ywm3JI76xAPn8xEmqEmgvI5zkrAd4D8=\",\"signature\":\"Q_WOcriLp8LOuWs4TsOy4cNp4bI1KtfWmA15mxnCNW3q0cbBN2q6dkyfHhm2ZvzdNuR89GbhWh1-yGUFk5lg0DPoMLHxg1Y1yJuBgv0ETb-G8_Wysu0GkWIJPN9mLKSWKI8X2ks6qBIN4FeW_KA2lhyduuZlwT9p9yzjBQokAL7aXtBtwS2L9kZeHuaGAwTqIfurwwFLeVa8R5-Zee-m2RtSbj4F-KjzAP0C3BuX44I6DrYoP3VDIMTEpbzQNkwPvH_NrrjRPIxeQl8VtKQsgb1iW4sC-24dD9jWetQZK3wO1awLOQNLomxiy4Db_E90-cAWQCuFKvFrSIqdo3mBrA==\",\"signatureAlgorithm\":\"rsa-sha256\",\"encryptionAlgorithm\":\"aes-cfb\"}"`)
	ds, err = PersistenceLoad(obj)
	log.Println(ds)
	log.Println(err)
	if err != nil {
		t.Error("Thrown error, got: ", err)
	}

	log.Println("\n\nIncorrect")
	session, _ = sm.GetSessionData(obj.ID)
	sha = utils.HashSUM256("qwerty12345")
	obj, _ = dto.PersistenceWithPasswordBuilder(obj.ID, session, sha)
	log.Println(session)
	obj.LocalFileBytes = []byte(`"{\"id\":\"DS_22414843-3426-41ff-a9c0-c6f72269ff5d\",\"encryptedData\":\"jO1dKmrfUbwck6ayMTQ0e8RitNNxx-nTh5NcBh_ZxMZkvcK3JfIjHnsHYTCZeIM6tU-f_dfgZn7kCglP1f_s642UO8LD4iEb6C-bb1i4S9MuP8RHQs1elrWNZohlrSE=\",\"signature\":\"PF4VUQQNHNd5renk17haaAUUk2ife29F_dJZY2lwJWgUjNYQqG9fCyD8sFzvsxtaWxNUbwnuwP5CTjcRt63ZxWxQJS29b8iXPYD6UX5SBmVRbmGNrwTnV7B9SSY-AIcsrq7iSIfpac3iE5MJ15O7vedjIR2t84tpGPU65Rl7dAncoR_UuxfFltuQ4D375RfvuStIcFiPs_dAiGXy6TUIQNdadCHHIR4LiK8SNjXX9jozAbZG9POdsCp2H6uwuHLLiiGIk0OQeJtTLmmTjadqvJ1BrHUgVmEvM_bhvgKwAijWGnzh1rB8RUU4Z6BFZIyXF6M86BlUc7dF4QBvqWUpLQ==\",\"signatureAlgorithm\":\"rsa-sha256\",\"encryptionAlgorithm\":\"aes-cfb\"}"	`)
	ds, err = PersistenceLoad(obj)
	log.Println(ds)
	log.Println(err)
	if err == nil {
		t.Error("Should have thrown error")
	}

	session, _ = sm.GetSessionData(obj.ID)
	sha = utils.HashSUM256("qwerty")
	obj, _ = dto.PersistenceWithPasswordBuilder(obj.ID, session, sha)
	log.Println(session)
	ds, err = BackChannelDecryption(obj, `"{\"id\":\"DS_521f98d6-7578-41bd-bbe3-7a3eeb35e1e1\",\"encryptedData\":\"a5zwnD-D1A1R3ZGlprOLqwKO7L_Ww0CRgqzqa4tkczRGqoCAPFmdZepv56HWDsGknfEUOXH_X0Am4tPQhurjS-X2Qq0yARP-Ywm3JI76xAPn8xEmqEmgvI5zkrAd4D8=\",\"signature\":\"Q_WOcriLp8LOuWs4TsOy4cNp4bI1KtfWmA15mxnCNW3q0cbBN2q6dkyfHhm2ZvzdNuR89GbhWh1-yGUFk5lg0DPoMLHxg1Y1yJuBgv0ETb-G8_Wysu0GkWIJPN9mLKSWKI8X2ks6qBIN4FeW_KA2lhyduuZlwT9p9yzjBQokAL7aXtBtwS2L9kZeHuaGAwTqIfurwwFLeVa8R5-Zee-m2RtSbj4F-KjzAP0C3BuX44I6DrYoP3VDIMTEpbzQNkwPvH_NrrjRPIxeQl8VtKQsgb1iW4sC-24dD9jWetQZK3wO1awLOQNLomxiy4Db_E90-cAWQCuFKvFrSIqdo3mBrA==\",\"signatureAlgorithm\":\"rsa-sha256\",\"encryptionAlgorithm\":\"aes-cfb\"}"`)
	log.Println(ds)
	log.Println("erro", err)
	if err != nil {
		t.Error("Thrown Error: ", err)
	}

	log.Println("\n\nIncorrect")
	session, _ = sm.GetSessionData(obj.ID)
	sha = utils.HashSUM256("qwerty")
	obj, _ = dto.PersistenceWithPasswordBuilder(obj.ID, session, sha)
	log.Println(session)
	ds, err = BackChannelDecryption(obj, `"{\"id\":\"DS_22414843-3426-41ff-a9c0-c6f72269ff5d\",\"encryptedData\":\"jO1dKmrfUbwck6ayMTQ0e8RitNNxx-nTh5NcBh_ZxMZkvcK3JfIjHnsHYTCZeIM6tU-f_dfgZn7kCglP1f_s642UO8LD4iEb6C-bb1i4S9MuP8RHQs1elrWNZohlrSE=\",\"signature\":\"PF4VsdfghjklUQQNHNd5renk17haaAUUk2ife29F_dJZY2lwJWgUjNYQqG9fCyD8sFzvsxtaWxNUbwnuwP5CTjcRt63ZxWxQJS29b8iXPYD6UX5SBmVRbmGNrwTnV7B9SSY-AIcsrq7iSIfpac3iE5MJ15O7vedjIR2t84tpGPU65Rl7dAncoR_UuxfFltuQ4D375RfvuStIcFiPs_dAiGXy6TUIQNdadCHHIR4LiK8SNjXX9jozAbZG9POdsCp2H6uwuHLLiiGIk0OQeJtTLmmTjadqvJ1BrHUgVmEvM_bhvgKwAijWGnzh1rB8RUU4Z6BFZIyXF6M86BlUc7dF4QBvqWUpLQ==\",\"signatureAlgorithm\":\"rsa-sha256\",\"encryptionAlgorithm\":\"aes-cfb\"}"`)
	log.Println(ds)
	log.Println(err)
	if err == nil {
		t.Error("Should have thrown error")
	}

	log.Println("\n\nIncorrect")
	session, _ = sm.GetSessionData(obj.ID)
	sha = utils.HashSUM256("qwerty")
	obj, _ = dto.PersistenceWithPasswordBuilder(obj.ID, session, sha)
	log.Println(session)
	obj.LocalFileBytes = []byte(`"{\"id12345678\":\"DS_22414843-3426-41ff-a9c0-c6f72269ff5d\",\"encryptedData\":\"jO1dKmrfUbwck6ayMTQ0e8RitNNxx-nTh5NcBh_ZxMZkvcK3JfIjHnsHYTCZeIM6tU-f_dfgZn7kCglP1f_s642UO8LD4iEb6C-bb1i4S9MuP8RHQs1elrWNZohlrSE=\",\"signature\":\"PF4VsdfghjklUQQNHNd5renk17haaAUUk2ife29F_dJZY2lwJWgUjNYQqG9fCyD8sFzvsxtaWxNUbwnuwP5CTjcRt63ZxWxQJS29b8iXPYD6UX5SBmVRbmGNrwTnV7B9SSY-AIcsrq7iSIfpac3iE5MJ15O7vedjIR2t84tpGPU65Rl7dAncoR_UuxfFltuQ4D375RfvuStIcFiPs_dAiGXy6TUIQNdadCHHIR4LiK8SNjXX9jozAbZG9POdsCp2H6uwuHLLiiGIk0OQeJtTLmmTjadqvJ1BrHUgVmEvM_bhvgKwAijWGnzh1rB8RUU4Z6BFZIyXF6M86BlUc7dF4QBvqWUpLQ==\",\"signatureAlgorithm\":\"rsa-sha256\",\"encryptionAlgorithm\":\"aes-cfb\"}"	`)
	ds, err = PersistenceLoad(obj)
	log.Println(ds)
	log.Println(err)
	if err == nil {
		t.Error("Should have thrown error")
	}

	session, _ = sm.GetSessionData(obj.ID)
	sha = utils.HashSUM256("qwerty")
	obj, _ = dto.PersistenceWithPasswordBuilder(obj.ID, session, sha)
	log.Println(session)
	obj.LocalFileBytes = []byte(`"{\"id12345678\":\"DS_22414843-3426-41ff-a9c0-c6f72269ff5d\",\"encryptedData\":\"jO1dKmrfUbwck6ayMTQ0e8RitNNxx-nTh5NcBh_ZxMZkvcK3JfIjHnsHYTCZeIM6tU-f_dfgZn7kCglP1f_s642UO8LD4iEb6C-bb1i4S9MuP8RHQs1elrWNZohlrSE=\",\"signature\":\"PF4VsdfghjklUQQNHNd5renk17haaAUUk2ife29F_dJZY2lwJWgUjNYQqG9fCyD8sFzvsxtaWxNUbwnuwP5CTjcRt63ZxWxQJS29b8iXPYD6UX5SBmVRbmGNrwTnV7B9SSY-AIcsrq7iSIfpac3iE5MJ15O7vedjIR2t84tpGPU65Rl7dAncoR_UuxfFltuQ4D375RfvuStIcFiPs_dAiGXy6TUIQNdadCHHIR4LiK8SNjXX9jozAbZG9POdsCp2H6uwuHLLiiGIk0OQeJtTLmmTjadqvJ1BrHUgVmEvM_bhvgKwAijWGnzh1rB8RUU4Z6BFZIyXF6M86BlUc7dF4QBvqWUpLQ==\",\"signatureAlgorithm\":\"rsa-sha256\",\"encryptionAlgorithm\":\"aes-cfb\"}"	`)
	response, err := BackChannelStorage(obj)
	log.Println(response)
	log.Println(err)
	if err != nil {
		t.Error("Thrown Error: ", err)
	}

}
