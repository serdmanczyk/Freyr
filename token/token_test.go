package token

import (
	"github.com/dgrijalva/jwt-go"
	"testing"
	"time"
)

var (
	testKey  = "tokenkeytokenkeytokenkey"
	badKey   = "bokenbokenboken"
	pkcs8key = "-----BEGIN PRIVATE KEY-----\nMIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQDRhQ57iLfj4LZ0\nqhj4vG5kg1RVF5uCaoJHfrG2mlL3SqMGGUoauosw3ONtXa9MBaoyci1VBBwzdGmP\nowwJKSwiIa9m8XoVaFvMxuTbcsQeljJGQ22bkgEDDSfs0vx5VxD9sSqqXeSRTxOz\nQSUBvykGX8kI4Sr5spOSdeRjGHCKjFqBfI771/XFLisSSIZBqJg9uk53t4K3oHC3\nQSCqZk+uBU1A5pnOpc+epoiAhennu1XvX5gZXXRspXXgLfehax6Q2pRr8BUkhdCX\nyj7gwlJ9HD9GXiMLSDSpyCCjB7Csu8A2Tsh2z4HT+fB9nnAclAdvIkHA69d6dFso\nqwob1E/9AgMBAAECggEBAJTbddiq4AuVAcyNdURzi/L5o5b5ONFFnt3w044qwrtT\nWdPhb9bhpjbHGQYAw6S9eZhxqmd2jhq4oK8eZlSz3dk6GYaIFfbTuDUbMkn/lfst\nNvbYvS0EZJeoZy9JD3ueMkIr92YnY1ch2ZtHs2U0TY9rykb2wzO4fkRWYjdNi9fC\navwoUOAESpaOcDckRyWz8RWldgtBpkhl0EDGWGWQ8Eo8uSiWKyUfE03N86Z/nQ7q\nFuY4l9r/E17V9b1jVGVZePrRGZ9yoxiiF+L4SzWRWhQvNKjghxdKme/g+m5oUczz\n/zCNn/doNmMF9/QShusYiHnJsnIoVsh7zDXxvx0UCbkCgYEA7CVFx1mMVvZwri7D\nyKGAOzSICHmCPrE+LWX8eiYei7NUu2zNljmVoPZ2Ou8SU+1uFBHmJD9qMG3aHzJq\n6cAtfHjWGLO/pXTvMqSudCh4nPfeHP9Shgx2CBHzW6CNj9JG5EWKbV+TDwPxo1eq\nX82ZaNT9uqcprqOc28YTsUl79DsCgYEA4yKygI2r2tSg4f02syseyuDOA0PIMrHg\n2303XFk7+oCd1euLii2nnjq/qXtlu7PST56aACRKKms6/V9BRL8hsLePMlFMHxar\nj8q1Ovf/BorPh9NyqzbeqLQNlazE93IxUgN0tVGrGnD1URWqX5p4XLkHniNi/oLA\nYUd8bwd7oScCgYEAxElbDhAuKh7gnLg8fylXGF9a74horcnQQBY03iePXlnrBXu0\nC7nD2S7kKaqNFnwV8tLJ9LlNgAHfu+zBl5jpdjxO4euPUm23YeYnKGB3mSojUwEb\nzFbRSXX6TeBPqwuDZ70yCiXWbDXABiEZelbAvLXGTf8jE4nmGXw05DmLsf8CgYBk\nLKFdYR4yXSS3ht3hF1t1TsCNYA+jjCAHraoE6LYzPRZfior4XjpW5sIxFWNA7YYL\n5380IM00+CYEKUa38zQApHRbVM+lxnHT8SsM3uNzFzWAShmAuapp7T8wjAoyuAJY\nkX2fmm1ENB19rXh+wbnj6xcY/7JhXXlLbiPLNBmqcQKBgQC/vBmBAxR0VN5UMQVw\nNav9vVXDj862PHNgLpTwuBkEf+OX8gjSa2OxKksy/v3xTkueh+FVEfrxZ5GDcjl/\n7Ei8DhfzsiVBBdbqKZpZvOCP8vi6gyNh4gemzvBxY4YYiy2aVGRqDQ6ra12Y8xWt\nO4cr2zZ5uhm+QoD//L3A8qD+YQ==\n-----END PRIVATE KEY-----"
)

func TestTokenGen(t *testing.T) {
	tkgen := JtwTokenGen(testKey)
	expiry := time.Now().Add(time.Minute)

	clams := Claims{"cheese": "swiss"}
	signedToken, err := tkgen.GenerateToken(expiry, clams)
	if err != nil {
		t.Fatalf("Failure generating token: %v", err)
	}

	parsedClaims, err := tkgen.ValidateToken(signedToken)
	if err != nil {
		t.Fatalf("Token verifaction failed: %v", err)
	}

	cheese, ok := parsedClaims["cheese"]
	if !ok {
		t.Fatalf("Claims not carried")
	}

	if cheese != "swiss" {
		t.Fatalf("Claim invalid, expected swiss, got: %s", cheese)
	}
}

func TestTokenExpiration(t *testing.T) {
	tkgen := JtwTokenGen(testKey)
	expiry := time.Now().Add(-time.Minute)
	signedToken, err := tkgen.GenerateToken(expiry, Claims{})
	if err != nil {
		t.Fatalf("Failure generating token: %v", err)
	}

	_, err = tkgen.ValidateToken(signedToken)
	if err != TokenExpired {
		t.Fatalf("Token should be expired")
	}
}

func TestKeyFailure(t *testing.T) {
	tkgen := JtwTokenGen(testKey)
	expiry := time.Now().Add(time.Minute)
	signedToken, err := tkgen.GenerateToken(expiry, Claims{})
	if err != nil {
		t.Fatalf("Failure generating token: %v", err)
	}

	otherTkgen := JtwTokenGen(badKey)

	_, err = otherTkgen.ValidateToken(signedToken)
	if err == nil || err.Error() != jwt.ErrSignatureInvalid.Error() {
		t.Fatalf("Signature should be invalid")
	}
}

func TestWrongAlgorithm(t *testing.T) {
	token := jwt.New(jwt.SigningMethodRS256)
	token.Claims["whatev"] = 1
	signedToken, err := token.SignedString([]byte(pkcs8key))
	if err != nil {
		t.Fatalf("Failure generating token: %v", err)
	}

	tkgen := JtwTokenGen(testKey)

	_, err = tkgen.ValidateToken(signedToken)
	if err == nil || err.Error() != InvalidAlgorithm.Error() {
		t.Fatalf("Algorithm should be invalid")
	}
}
