package services

import (
	"fmt"
	"testing"

	"github.com/EC-SEAL/perseal/sm"
	"github.com/EC-SEAL/perseal/utils"
)

func TestStoreService(t *testing.T) {

	var passed = "=================PASSED==============="
	var failed = "=================FAILED==============="

	obj := InitIntegration("Browser")
	smResp, _ := sm.GetSessionData(obj.ID)

	fmt.Println("\n=================Correct Persistence Store====================")
	obj = preCloudConfig(obj, smResp, "qwerty")
	_, erro := PersistenceStore(obj)
	if erro != nil {
		fmt.Println(failed)
		t.Error("Thrown error, got: ", erro)
	} else {
		fmt.Println(passed)
	}

	fmt.Println("\n=================Correct Persistence Load====================")
	sha := utils.HashSUM256("qwerty")
	obj = preCloudConfig(obj, smResp, sha)
	obj.LocalFileBytes = []byte(`"{\"id\":\"db6a4657-c419-4d97-981c-a2dc6c984c82\",\"encryptedData\":\"R9wD2odntxKInqnFXRH-2_TfHUDi8K9aK5fTkw5U_wqUHkG8Carlx6QNnDqYJQe7MdFfy_d9z-aH28ftqCo2-huT5v2m8kp6vjXQnd5ufd802jxT9LsW_A5Te22bxrgL6yso1zTDODA1IlUPG86_nx2MLxtDCfr9yZq1fnmg0TfsRQ==\",\"signature\":\"H3YAEst4CXSHY9oVlIauhkDKFpm4WJOm_0Yr0py461VCNwr3QvLlROavkkq532WQNhPWyJWXumOsmlHuLqGQuf8lOWz4EXXqaBbYSXefB9z7IhH8FWIeyOqxp_C0kD2mOILQnmvO91i9oCse6XbBcPgh7IwfM_uP6bmMa_pUbxzT7dYT2WV93EshCNEHv4MSO1-8RnhUtFZEkq1lKbwDr3uSkKMu0Tmhgo3K8CkZ3BfgbC3-vLY9MFtq69bVozK07FVSlAxFGgmAW21lR6gNSKObYcQyYttkVMDTd9MrnTg-E7Qhsf41K8WVY6-ea6NXhzWqB4CcYAkRfteSF_hT9Q==\",\"signatureAlgorithm\":\"rsa-sha256\",\"encryptionAlgorithm\":\"aes-cfb\"}"
	`)
	_, err := PersistenceLoad(obj)
	if err != nil {
		fmt.Println(failed)
		t.Error("Thrown error, got: ", err)
	} else {
		fmt.Println(passed)
	}

	fmt.Println("\n=================Incorrect Persistence Load - Bad Password====================")
	sha = utils.HashSUM256("qwerty12345")
	obj = preCloudConfig(obj, smResp, sha)
	obj.LocalFileBytes = []byte(`"{\"id\":\"DS_22414843-3426-41ff-a9c0-c6f72269ff5d\",\"encryptedData\":\"jO1dKmrfUbwck6ayMTQ0e8RitNNxx-nTh5NcBh_ZxMZkvcK3JfIjHnsHYTCZeIM6tU-f_dfgZn7kCglP1f_s642UO8LD4iEb6C-bb1i4S9MuP8RHQs1elrWNZohlrSE=\",\"signature\":\"PF4VUQQNHNd5renk17haaAUUk2ife29F_dJZY2lwJWgUjNYQqG9fCyD8sFzvsxtaWxNUbwnuwP5CTjcRt63ZxWxQJS29b8iXPYD6UX5SBmVRbmGNrwTnV7B9SSY-AIcsrq7iSIfpac3iE5MJ15O7vedjIR2t84tpGPU65Rl7dAncoR_UuxfFltuQ4D375RfvuStIcFiPs_dAiGXy6TUIQNdadCHHIR4LiK8SNjXX9jozAbZG9POdsCp2H6uwuHLLiiGIk0OQeJtTLmmTjadqvJ1BrHUgVmEvM_bhvgKwAijWGnzh1rB8RUU4Z6BFZIyXF6M86BlUc7dF4QBvqWUpLQ==\",\"signatureAlgorithm\":\"rsa-sha256\",\"encryptionAlgorithm\":\"aes-cfb\"}"	`)
	_, err = PersistenceLoad(obj)
	if err == nil {
		fmt.Println(failed)
		t.Error("Should have thrown error")
	} else {
		fmt.Println(passed)
	}

	fmt.Println("\n=================Incorrect Persistence Load - Bad File====================")
	obj = preCloudConfig(obj, smResp, sha)
	obj.LocalFileBytes = []byte(`"{\"id12345678\":\"DS_22414843-3426-41ff-a9c0-c6f72269ff5d\",\"encryptedData\":\"jO1dKmrfUbwck6ayMTQ0e8RitNNxx-nTh5NcBh_ZxMZkvcK3JfIjHnsHYTCZeIM6tU-f_dfgZn7kCglP1f_s642UO8LD4iEb6C-bb1i4S9MuP8RHQs1elrWNZohlrSE=\",\"signature\":\"PF4VsdfghjklUQQNHNd5renk17haaAUUk2ife29F_dJZY2lwJWgUjNYQqG9fCyD8sFzvsxtaWxNUbwnuwP5CTjcRt63ZxWxQJS29b8iXPYD6UX5SBmVRbmGNrwTnV7B9SSY-AIcsrq7iSIfpac3iE5MJ15O7vedjIR2t84tpGPU65Rl7dAncoR_UuxfFltuQ4D375RfvuStIcFiPs_dAiGXy6TUIQNdadCHHIR4LiK8SNjXX9jozAbZG9POdsCp2H6uwuHLLiiGIk0OQeJtTLmmTjadqvJ1BrHUgVmEvM_bhvgKwAijWGnzh1rB8RUU4Z6BFZIyXF6M86BlUc7dF4QBvqWUpLQ==\",\"signatureAlgorithm\":\"rsa-sha256\",\"encryptionAlgorithm\":\"aes-cfb\"}"	`)
	_, err = PersistenceLoad(obj)
	if err == nil {
		fmt.Println(failed)
		t.Error("Should have thrown error")
	} else {
		fmt.Println(passed)
	}

	fmt.Println("\n=================Correct Back-Channel Decryption====================")
	sha = utils.HashSUM256("qwerty")
	obj = preCloudConfig(obj, smResp, sha)
	obj.LocalFileBytes = []byte(`"{\"id\":\"ea0f1b47-dad6-446c-8969-1353602a97b1\",\"encryptedData\":\"rwOQtGwIeUTLWAKsURoLWFWV0IE=\",\"signature\":\"EKq9SDnKEZijD9IGZGxS8EP5uqTZQ-aY_rLp1iiQUDhSYy9MzUstCyI0ryOuUgHrYDrhpR73ZWV7tJsZ8fxJKqXpbGLx1_i-pX6AeyAwPIBGizQl4sbBqN2OMNKSVvVztnzbjdWzAnqIM2IXgmWobnR8BoYZupGremLx25t_nzoNp8EGpwNk_DlmkyFLJTCBIJLyoJ_6-WG49V2A--32WcBXpVq939Q0r44zxGrCBMLH742vCQe8lRAn8YeQinYo0U7Kuc7sIS02zdCoPiRLsVcP6d7N9iNBOxT_iEsGj57y16ACHgbcooQCWhBQlv1wdphsBFVYr0xq71_1XISGhQ==\",\"signatureAlgorithm\":\"rsa-sha256\",\"encryptionAlgorithm\":\"aes-cfb\",\"clearData\":\"[]\"}"`)
	response, err := BackChannelDecryption(obj, string(obj.LocalFileBytes))
	if response.Code != 200 {
		fmt.Println(failed)
		t.Error("Bad Response")
	}
	if err != nil {
		fmt.Println(failed)
		t.Error("Thrown error, got: ", err)
	} else {
		fmt.Println(passed)
	}

	fmt.Println("\n=================Incorrect Back-Channel Decryption - Bad File====================")
	obj = preCloudConfig(obj, smResp, sha)
	_, err = BackChannelDecryption(obj, `"{\"id\":\"DS_22414843-3426-41ff-a9c0-c6f72269ff5d\",\"encryptedData\":\"jO1dKmrfUbwck6ayMTQ0e8RitNNxx-nTh5NcBh_ZxMZkvcK3JfIjHnsHYTCZeIM6tU-f_dfgZn7kCglP1f_s642UO8LD4iEb6C-bb1i4S9MuP8RHQs1elrWNZohlrSE=\",\"signature\":\"PF4VsdfghjklUQQNHNd5renk17haaAUUk2ife29F_dJZY2lwJWgUjNYQqG9fCyD8sFzvsxtaWxNUbwnuwP5CTjcRt63ZxWxQJS29b8iXPYD6UX5SBmVRbmGNrwTnV7B9SSY-AIcsrq7iSIfpac3iE5MJ15O7vedjIR2t84tpGPU65Rl7dAncoR_UuxfFltuQ4D375RfvuStIcFiPs_dAiGXy6TUIQNdadCHHIR4LiK8SNjXX9jozAbZG9POdsCp2H6uwuHLLiiGIk0OQeJtTLmmTjadqvJ1BrHUgVmEvM_bhvgKwAijWGnzh1rB8RUU4Z6BFZIyXF6M86BlUc7dF4QBvqWUpLQ==\",\"signatureAlgorithm\":\"rsa-sha256\",\"encryptionAlgorithm\":\"aes-cfb\"}"`)
	if err == nil {
		fmt.Println(failed)
		t.Error("Should have thrown error")
	} else {
		fmt.Println(passed)
	}

	fmt.Println("\n=================Incorrect Back-Channel Decryption - Bad Password====================")
	obj = preCloudConfig(obj, smResp, sha+"12")
	obj.LocalFileBytes = []byte(`"{\"id\":\"ea0f1b47-dad6-446c-8969-1353602a97b1\",\"encryptedData\":\"rwOQtGwIeUTLWAKsURoLWFWV0IE=\",\"signature\":\"EKq9SDnKEZijD9IGZGxS8EP5uqTZQ-aY_rLp1iiQUDhSYy9MzUstCyI0ryOuUgHrYDrhpR73ZWV7tJsZ8fxJKqXpbGLx1_i-pX6AeyAwPIBGizQl4sbBqN2OMNKSVvVztnzbjdWzAnqIM2IXgmWobnR8BoYZupGremLx25t_nzoNp8EGpwNk_DlmkyFLJTCBIJLyoJ_6-WG49V2A--32WcBXpVq939Q0r44zxGrCBMLH742vCQe8lRAn8YeQinYo0U7Kuc7sIS02zdCoPiRLsVcP6d7N9iNBOxT_iEsGj57y16ACHgbcooQCWhBQlv1wdphsBFVYr0xq71_1XISGhQ==\",\"signatureAlgorithm\":\"rsa-sha256\",\"encryptionAlgorithm\":\"aes-cfb\",\"clearData\":\"[]\"}"`)
	response, err = BackChannelDecryption(obj, string(obj.LocalFileBytes))
	if err == nil {
		fmt.Println(failed)
		t.Error("Should have thrown error")
	} else {
		fmt.Println(passed)
	}

	fmt.Println("\n=================Incorrect Back-Channel Storage - Bad File====================")
	obj = preCloudConfig(obj, smResp, sha)
	obj.LocalFileBytes = []byte(`"{\"id12345678\":\"DS_22414843-3426-41ff-a9c0-c6f72269ff5d\",\"encryptedData\":\"jO1dKmrfUbwck6ayMTQ0e8RitNNxx-nTh5NcBh_ZxMZkvcK3JfIjHnsHYTCZeIM6tU-f_dfgZn7kCglP1f_s642UO8LD4iEb6C-bb1i4S9MuP8RHQs1elrWNZohlrSE=\",\"signature\":\"PF4VsdfghjklUQQNHNd5renk17haaAUUk2ife29F_dJZY2lwJWgUjNYQqG9fCyD8sFzvsxtaWxNUbwnuwP5CTjcRt63ZxWxQJS29b8iXPYD6UX5SBmVRbmGNrwTnV7B9SSY-AIcsrq7iSIfpac3iE5MJ15O7vedjIR2t84tpGPU65Rl7dAncoR_UuxfFltuQ4D375RfvuStIcFiPs_dAiGXy6TUIQNdadCHHIR4LiK8SNjXX9jozAbZG9POdsCp2H6uwuHLLiiGIk0OQeJtTLmmTjadqvJ1BrHUgVmEvM_bhvgKwAijWGnzh1rB8RUU4Z6BFZIyXF6M86BlUc7dF4QBvqWUpLQ==\",\"signatureAlgorithm\":\"rsa-sha256\",\"encryptionAlgorithm\":\"aes-cfb\"}"	`)
	response, err = BackChannelStorage(obj)
	if err != nil {
		fmt.Println(failed)
		t.Error("Thrown error, got: ", err)
	} else {
		fmt.Println(passed)
	}

	fmt.Println("\n=================Correct Persistence & Store Browser Load====================")
	sha = utils.HashSUM256("qwerty")
	obj.Method = "Browser"
	obj = preCloudConfig(obj, smResp, sha)
	obj.LocalFileBytes = []byte(`"{\"id\":\"db6a4657-c419-4d97-981c-a2dc6c984c82\",\"encryptedData\":\"R9wD2odntxKInqnFXRH-2_TfHUDi8K9aK5fTkw5U_wqUHkG8Carlx6QNnDqYJQe7MdFfy_d9z-aH28ftqCo2-huT5v2m8kp6vjXQnd5ufd802jxT9LsW_A5Te22bxrgL6yso1zTDODA1IlUPG86_nx2MLxtDCfr9yZq1fnmg0TfsRQ==\",\"signature\":\"H3YAEst4CXSHY9oVlIauhkDKFpm4WJOm_0Yr0py461VCNwr3QvLlROavkkq532WQNhPWyJWXumOsmlHuLqGQuf8lOWz4EXXqaBbYSXefB9z7IhH8FWIeyOqxp_C0kD2mOILQnmvO91i9oCse6XbBcPgh7IwfM_uP6bmMa_pUbxzT7dYT2WV93EshCNEHv4MSO1-8RnhUtFZEkq1lKbwDr3uSkKMu0Tmhgo3K8CkZ3BfgbC3-vLY9MFtq69bVozK07FVSlAxFGgmAW21lR6gNSKObYcQyYttkVMDTd9MrnTg-E7Qhsf41K8WVY6-ea6NXhzWqB4CcYAkRfteSF_hT9Q==\",\"signatureAlgorithm\":\"rsa-sha256\",\"encryptionAlgorithm\":\"aes-cfb\"}"
	`)
	_, err = PersistenceStoreAndLoad(obj)
	if err != nil {
		fmt.Println(failed)
		t.Error("Thrown error, got: ", err)
	} else {
		fmt.Println(passed)
	}

	fmt.Println("\n=================Incorrect Persistence Browser Load====================")
	sha = utils.HashSUM256("qwerty")
	obj.Method = "Browser"
	obj = preCloudConfig(obj, smResp, sha)
	obj.LocalFileBytes = []byte(`"{\"id\":\"db6a4657-c419-4d97-981c-a2dc6cptedData\":\"R9wD2odntxKInqnFXRH-2_TfHUDi8K9aK5fTkw5U_wqUHkG8Carlx6QNnDqYJQe7MdFfy_d9z-aH28ftqCo2-huT5v2m8kp6vjXQnd5ufd802jxT9LsW_A5Te22bxrgL6yso1zTDODA1IlUPG86_nx2MLxtDCfr9yZq1fnmg0TfsRQ==\",\"signature\":\"H3YAEst4CXSHY9oVlIauhkDKFpm4WJOm_0Yr0py461VCNwr3QvLlROavkkq532WQNhPWyJWXumOsmlHuLqGQuf8lOWz4EXXqaBbYSXefB9z7IhH8FWIeyOqxp_C0kD2mOILQnmvO91i9oCse6XbBcPgh7IwfM_uP6bmMa_pUbxzT7dYT2WV93EshCNEHv4MSO1-8RnhUtFZEkq1lKbwDr3uSkKMu0Tmhgo3K8CkZ3BfgbC3-vLY9MFtq69bVozK07FVSlAxFGgmAW21lR6gNSKObYcQyYttkVMDTd9MrnTg-E7Qhsf41K8WVY6-ea6NXhzWqB4CcYAkRfteSF_hT9Q==\",\"signatureAlgorithm\":\"rsa-sha256\",\"encryptionAlgorithm\":\"aes-cfb\"}"
	`)
	_, err = PersistenceLoad(obj)
	if err == nil {
		fmt.Println(failed)
		t.Error("Should have thrown error")
	} else {
		fmt.Println(passed)
	}

}
