package kustomize

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildSopsEncrypted(t *testing.T) {
	tempDir := t.TempDir()
	err := os.WriteFile(path.Join(tempDir, "secret.yaml"), []byte(`
apiVersion: v1
kind: Secret
metadata:
    name: redfish-creds-k8s-nodes
    annotations:
        metalkast.io/redfish-urls: |-
            https://192.168.122.101
            https://192.168.122.102
            https://192.168.122.103
data:
    username: ENC[AES256_GCM,data:FLKok0nSVCE=,iv:hvwgyk48JLBeAT9bUaJcdhaaLD8yVRAnKYA1IPnYNI0=,tag:2HkUei3vINmeS3R+1k0wfQ==,type:str]
    password: ENC[AES256_GCM,data:K9BauC1/p9eCsHHL,iv:sfMXRS/+D+ZVtpMK4YHeYXnD8NwmfzupstwQFzQj0No=,tag:Vc3+uvItiSU+xBqjRkx0fQ==,type:str]
type: Opaque
sops:
    kms: []
    gcp_kms: []
    azure_kv: []
    hc_vault: []
    age:
        - recipient: age1dk25phcnxzhkryzn7smn29wa4lhsplgvty3skzddr2w5plsh0ddq04ukv3
          enc: |
            -----BEGIN AGE ENCRYPTED FILE-----
            YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+IFgyNTUxOSA5RHJrUTAzdzhlZWNWM2lr
            dHFSc2o1dE53aTdXRldDcm9iWVdibHRueVJRCncxREpmejFESTB1d1YxQkh5c2Vu
            TUM5bmh4NXd3Yjl1SUVwcUtYYml5ME0KLS0tIG0wMS9SNitoMUkvZ0g5d2dVcWVu
            bG5GWHZkZXBHclZmREpGbjRidjgzaDAKvDAV+HePfd5UcsRm8KyxHkn4YCAJVkFK
            W3Dq8cEOE7yqWww7Uv7RxXbFO+C+3qS7tCjADEShc7cIfk+Z9QTmSg==
            -----END AGE ENCRYPTED FILE-----
    lastmodified: "2023-08-27T20:00:02Z"
    mac: ENC[AES256_GCM,data:iE5v36JugQgA2MqwaG018G3LRUu5EcMsRrimsDsx+fwhsXZpJ3hEq4fJd4SFbTBg21rfOLKT7KHDVf9iO9ISj476NYnItAuStKOGCYKNZaJzwKayonSCJNa54mtuiKrf+HD2haHm65SjvHRcAqlkynafg2JQmRJocQOJU+oHbS8=,iv:LCFqUyzmUjVK1DxueByXwQhLYhKmQUqSfVSat8KmtHg=,tag:zy8sBioWpFwg80Nzt0CUDA==,type:str]
    pgp: []
    encrypted_regex: ^(data|stringData)$
    version: 3.7.3
`), 0600)
	assert.NoError(t, err)

	err = os.WriteFile(path.Join(tempDir, "multiple-secrets.yaml"), []byte(`
apiVersion: v1
kind: Secret
metadata:
    name: multi-doc-1
stringData:
    foo1: ENC[AES256_GCM,data:iKd+uQ==,iv:jVYWsDEldIExqfBwbqYp9f/b+c3YEz2SNxzCNvDGON8=,tag:VJ6Z6ZFUT/zbKcZs9i4riw==,type:str]
sops:
    kms: []
    gcp_kms: []
    azure_kv: []
    hc_vault: []
    age:
        - recipient: age1dk25phcnxzhkryzn7smn29wa4lhsplgvty3skzddr2w5plsh0ddq04ukv3
          enc: |
            -----BEGIN AGE ENCRYPTED FILE-----
            YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+IFgyNTUxOSBvOVgxdnhvU3dvRkQyb3lu
            SXpqVERMMzk0eURFVlVWKzNzNGFQM2pFRjJBCmJBZmxxcHFLRDNiK2pYb3YxM0dO
            dHA1Y3ZSUVhodUs0eEtTV3RNSkxhcW8KLS0tIC81ZHBkTTVCQm13V0hyUzVxbWQ5
            bGtBc0x4QXcyVWVjd1RLNUo1cDhCVEEKGbVCKvxZUH/VT9G3wQSsVAp3JxQhYujk
            omFImYSplMqRSJ1NVChSHYCl4IiayO93stSaoEK1S1V+TAWGD37vfw==
            -----END AGE ENCRYPTED FILE-----
    lastmodified: "2023-08-30T17:30:55Z"
    mac: ENC[AES256_GCM,data:0Blc1S28cHZsmlLnHEYeSIinS12+I0x/+/x37SnNqLOhapdG1RfzwioJ1jdf+pIRipE1yAYjSlWCQ7ReQQZeubW0gru7JpgPen+jD5jT/k6FVeaH4wnvIid68SGKjBRed13jE79EhF0aokATFz3W9V4tmV9q152DK2EyftUBpaI=,iv:tFY/SWrvbnCnh4xE57qTM/n85hQE12kXoZ4GyzUxvIE=,tag:Ral49yMeFdFOgeA3ZH/8+Q==,type:str]
    pgp: []
    encrypted_regex: ^(data|stringData)$
    version: 3.7.3
---
apiVersion: v1
kind: Secret
metadata:
    name: multi-doc-2
stringData:
    foo2: ENC[AES256_GCM,data:14pQ2g==,iv:MUsidqBh0nXcP/QHLU6MagPpd6eRO0gymiSodkU7JxY=,tag:1laKlj4ovIm0ZcbOoP7wFg==,type:str]
sops:
    kms: []
    gcp_kms: []
    azure_kv: []
    hc_vault: []
    age:
        - recipient: age1dk25phcnxzhkryzn7smn29wa4lhsplgvty3skzddr2w5plsh0ddq04ukv3
          enc: |
            -----BEGIN AGE ENCRYPTED FILE-----
            YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+IFgyNTUxOSBvOVgxdnhvU3dvRkQyb3lu
            SXpqVERMMzk0eURFVlVWKzNzNGFQM2pFRjJBCmJBZmxxcHFLRDNiK2pYb3YxM0dO
            dHA1Y3ZSUVhodUs0eEtTV3RNSkxhcW8KLS0tIC81ZHBkTTVCQm13V0hyUzVxbWQ5
            bGtBc0x4QXcyVWVjd1RLNUo1cDhCVEEKGbVCKvxZUH/VT9G3wQSsVAp3JxQhYujk
            omFImYSplMqRSJ1NVChSHYCl4IiayO93stSaoEK1S1V+TAWGD37vfw==
            -----END AGE ENCRYPTED FILE-----
    lastmodified: "2023-08-30T17:30:55Z"
    mac: ENC[AES256_GCM,data:0Blc1S28cHZsmlLnHEYeSIinS12+I0x/+/x37SnNqLOhapdG1RfzwioJ1jdf+pIRipE1yAYjSlWCQ7ReQQZeubW0gru7JpgPen+jD5jT/k6FVeaH4wnvIid68SGKjBRed13jE79EhF0aokATFz3W9V4tmV9q152DK2EyftUBpaI=,iv:tFY/SWrvbnCnh4xE57qTM/n85hQE12kXoZ4GyzUxvIE=,tag:Ral49yMeFdFOgeA3ZH/8+Q==,type:str]
    pgp: []
    encrypted_regex: ^(data|stringData)$
    version: 3.7.3
`), 0644)
	assert.NoError(t, err)

	err = os.WriteFile(path.Join(tempDir, "kustomization.yaml"), []byte(`
resources:
- secret.yaml
- multiple-secrets.yaml

configMapGenerator:
- name: foo
  options:
    disableNameSuffixHash: true
  envs:
  - envs-yaml-parsable.txt
  - envs-non-yaml-parsable.txt

namespace: foo
    `), 0644)
	assert.NoError(t, err)

	// Parsed as YAML string
	err = os.WriteFile(path.Join(tempDir, "envs-yaml-parsable.txt"), []byte(`
FOO=BAR
    `), 0644)
	assert.NoError(t, err)

	permissionDeniedFile := path.Join(tempDir, "envs-non-yaml-parsable.txt")
	err = os.WriteFile(permissionDeniedFile, []byte(`
FOO2=BAR2:
FOO3=BAR3
    `), 0000)
	assert.NoError(t, err)

	want := `apiVersion: v1
data:
  FOO: BAR
  FOO2: 'BAR2:'
  FOO3: BAR3
kind: ConfigMap
metadata:
  name: foo
  namespace: foo
---
apiVersion: v1
kind: Secret
metadata:
  name: multi-doc-1
  namespace: foo
stringData:
  foo1: bar1
---
apiVersion: v1
kind: Secret
metadata:
  name: multi-doc-2
  namespace: foo
stringData:
  foo2: bar2
---
apiVersion: v1
data:
  password: cGFzc3dvcmQ=
  username: YWRtaW4=
kind: Secret
metadata:
  annotations:
    metalkast.io/redfish-urls: |-
      https://192.168.122.101
      https://192.168.122.102
      https://192.168.122.103
  name: redfish-creds-k8s-nodes
  namespace: foo
type: Opaque
`

	// No decryption key configured
	_, err = Build(tempDir)
	assert.Error(t, err)

	t.Setenv("SOPS_AGE_KEY", "AGE-SECRET-KEY-15JLZDHADZ45JVZXMKSAM9U8AHE47DDK7DTJL7XNR0G27U4P9XRHQLNKUH2")

	// file cannot be read: permission denied
	_, err = Build(tempDir)
	assert.Error(t, err)

	// fix permissions
	err = os.Chmod(permissionDeniedFile, 0644)
	assert.NoError(t, err)

	result, err := Build(tempDir)
	assert.NoError(t, err)
	assert.Equal(t, want, string(result))
}
