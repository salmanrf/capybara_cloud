package auth

import "github.com/alexedwards/argon2id"

func Hash(pwd string) (def string, err error) {
	hash, err := argon2id.CreateHash(pwd, argon2id.DefaultParams)

	if err != nil {
		return def, err
	}

	return hash, nil
}

func HashCompare(pwd, hash string) (def bool, err error) {
	match, err := argon2id.ComparePasswordAndHash(pwd, hash)

	if err != nil {
		return def, err
	}

	return match, nil
}