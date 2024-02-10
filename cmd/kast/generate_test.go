package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateBareMetalHosts(t *testing.T) {
	input := `
apiVersion: v1
kind: Secret
metadata:
  name: k8s-nodes-a
  annotations:
    metalkast.io/redfish-urls: |-
      http://host1/
      http://host2
stringData:
  username: foo1
  password: bar1
type: Opaque
---
apiVersion: v1
kind: Secret
metadata:
  name: k8s-nodes-b
  namespace: test-ns
  annotations:
    metalkast.io/redfish-urls: |-
      http://host3:9000/
      http://host4:9000
data:
  username: Zm9vMg==
  password: YmFyMg==
type: Opaque
---
apiVersion: v1
kind: Secret
metadata:
  name: k8s-nodes-empty
  namespace: test-ns
  annotations:
    metalkast.io/redfish-urls:
stringData:
  username: foo3
  password: bar3
type: Opaque
---
apiVersion: v1
kind: Secret
metadata:
  name: unreleated-secret
stringData:
  username: foo4
  password: bar4
type: Opaque
`

	inputEncrypted := `
apiVersion: v1
kind: Secret
metadata:
    name: k8s-nodes-a
    annotations:
        metalkast.io/redfish-urls: |-
            http://host1/
            http://host2
stringData:
    username: ENC[AES256_GCM,data:vGJkxg==,iv:GqJMfMCfFlFYBXWQs4AcpE0JvEJp5HtpO5ExvHWALrk=,tag:+0Z6bR06alIauVR3Mp1o0A==,type:str]
    password: ENC[AES256_GCM,data:gZt3Iw==,iv:GLdNb4FC5gx3DEMB0sc3Zr0/HVtOcj95K3Id6blrxV0=,tag:7C3hDoSxICWhLGYyXtNCcQ==,type:str]
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
            YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+IFgyNTUxOSA4WmRCSmdScDZja3AxZ3pE
            VGR2TWdCdGFpajFmYkR3TkVHWXJ0MUFleUhrCktFNWs5eVhoQUowOU9DVS8zbnh4
            SExFZGs5eW5ETHFjUUlrRGFlSk02cTQKLS0tIHI0L2FwelN2OHlkUjBHamh2NDBo
            am5iOUZHUEJacXlqRGpaRHJXalY4Nk0KrBCqOuUeZXtffJ4bLZt6i4KRIqiAFMa3
            7GFFkVCAYTsK1ZfP6vM4hUKvNe43n+GeAJbRJbpo11lccKVR1TdeyA==
            -----END AGE ENCRYPTED FILE-----
    lastmodified: "2023-08-27T19:02:02Z"
    mac: ENC[AES256_GCM,data:AgH0gLKrH537fjevKuiebwA9kw52Tl/TWSH6HctLcyiO8YHwsMJtQRZxHPKhJlNkFRV8BHjExfMgKEijOeMFfmOeF0yG73/3vk8o5hVtDIj2d8J4pLWvXeA7FU4/js8Z4u8JTvjewap4jHVMDlrlVGOTthJZ9ahW/hLF3EoCZtI=,iv:/rDngTcPDndx2UEj/5c3au1JPIiSWle2Rrd7F+KlbH0=,tag:9xI0inYNWbqTyA6ob61Qyg==,type:str]
    pgp: []
    encrypted_regex: ^(data|stringData)$
    version: 3.7.3
---
apiVersion: v1
kind: Secret
metadata:
    name: k8s-nodes-b
    namespace: test-ns
    annotations:
        metalkast.io/redfish-urls: |-
            http://host3:9000/
            http://host4:9000
data:
    username: ENC[AES256_GCM,data:lOFTwjscUgo=,iv:nIUJLcxju8hG6vOZkaLP1GRDb53mtyqRLdwNusHZZKs=,tag:/WSb3uEpCXRhLq4Lwi6lXQ==,type:str]
    password: ENC[AES256_GCM,data:ofsJZrr5DpI=,iv:ViUiHrw6RnS+YMuGLVDYBKME1i1VhFaaig61rIkMHXA=,tag:EZaetn4nMtYFIwtQ/VH6Rg==,type:str]
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
            YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+IFgyNTUxOSA4WmRCSmdScDZja3AxZ3pE
            VGR2TWdCdGFpajFmYkR3TkVHWXJ0MUFleUhrCktFNWs5eVhoQUowOU9DVS8zbnh4
            SExFZGs5eW5ETHFjUUlrRGFlSk02cTQKLS0tIHI0L2FwelN2OHlkUjBHamh2NDBo
            am5iOUZHUEJacXlqRGpaRHJXalY4Nk0KrBCqOuUeZXtffJ4bLZt6i4KRIqiAFMa3
            7GFFkVCAYTsK1ZfP6vM4hUKvNe43n+GeAJbRJbpo11lccKVR1TdeyA==
            -----END AGE ENCRYPTED FILE-----
    lastmodified: "2023-08-27T19:02:02Z"
    mac: ENC[AES256_GCM,data:AgH0gLKrH537fjevKuiebwA9kw52Tl/TWSH6HctLcyiO8YHwsMJtQRZxHPKhJlNkFRV8BHjExfMgKEijOeMFfmOeF0yG73/3vk8o5hVtDIj2d8J4pLWvXeA7FU4/js8Z4u8JTvjewap4jHVMDlrlVGOTthJZ9ahW/hLF3EoCZtI=,iv:/rDngTcPDndx2UEj/5c3au1JPIiSWle2Rrd7F+KlbH0=,tag:9xI0inYNWbqTyA6ob61Qyg==,type:str]
    pgp: []
    encrypted_regex: ^(data|stringData)$
    version: 3.7.3
---
apiVersion: v1
kind: Secret
metadata:
    name: k8s-nodes-empty
    namespace: test-ns
    annotations:
        metalkast.io/redfish-urls: null
stringData:
    username: ENC[AES256_GCM,data:pd2rcA==,iv:GiqI9J0rAPsgelBCB2W5iOi5R/fECUMEg3u2c5/EI10=,tag:tR+U8lvqHOXQxETtc1HezQ==,type:str]
    password: ENC[AES256_GCM,data:1dc5YA==,iv:KDE1zVNJMszVaZzQsalqYaf5fFKcoDZR13qFRe6xla8=,tag:UWCT9kNHuluoMt/4sScAqA==,type:str]
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
            YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+IFgyNTUxOSA4WmRCSmdScDZja3AxZ3pE
            VGR2TWdCdGFpajFmYkR3TkVHWXJ0MUFleUhrCktFNWs5eVhoQUowOU9DVS8zbnh4
            SExFZGs5eW5ETHFjUUlrRGFlSk02cTQKLS0tIHI0L2FwelN2OHlkUjBHamh2NDBo
            am5iOUZHUEJacXlqRGpaRHJXalY4Nk0KrBCqOuUeZXtffJ4bLZt6i4KRIqiAFMa3
            7GFFkVCAYTsK1ZfP6vM4hUKvNe43n+GeAJbRJbpo11lccKVR1TdeyA==
            -----END AGE ENCRYPTED FILE-----
    lastmodified: "2023-08-27T19:02:02Z"
    mac: ENC[AES256_GCM,data:AgH0gLKrH537fjevKuiebwA9kw52Tl/TWSH6HctLcyiO8YHwsMJtQRZxHPKhJlNkFRV8BHjExfMgKEijOeMFfmOeF0yG73/3vk8o5hVtDIj2d8J4pLWvXeA7FU4/js8Z4u8JTvjewap4jHVMDlrlVGOTthJZ9ahW/hLF3EoCZtI=,iv:/rDngTcPDndx2UEj/5c3au1JPIiSWle2Rrd7F+KlbH0=,tag:9xI0inYNWbqTyA6ob61Qyg==,type:str]
    pgp: []
    encrypted_regex: ^(data|stringData)$
    version: 3.7.3
---
apiVersion: v1
kind: Secret
metadata:
    name: unreleated-secret
stringData:
    username: ENC[AES256_GCM,data:dTa42Q==,iv:YU1c6a5Zayu+L51vfpCKO4nIKZ7Rjpa8Z7As1udtKgA=,tag:BNoJ0B2HoA5k1iKrxUnpuA==,type:str]
    password: ENC[AES256_GCM,data:OHTTyg==,iv:VijpQDmd7kpYcs/Fj5aGQL5DaCu80ysI5l3yB0Y4unE=,tag:7CQr+3WXRtAiyO3M8BEONg==,type:str]
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
            YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+IFgyNTUxOSA4WmRCSmdScDZja3AxZ3pE
            VGR2TWdCdGFpajFmYkR3TkVHWXJ0MUFleUhrCktFNWs5eVhoQUowOU9DVS8zbnh4
            SExFZGs5eW5ETHFjUUlrRGFlSk02cTQKLS0tIHI0L2FwelN2OHlkUjBHamh2NDBo
            am5iOUZHUEJacXlqRGpaRHJXalY4Nk0KrBCqOuUeZXtffJ4bLZt6i4KRIqiAFMa3
            7GFFkVCAYTsK1ZfP6vM4hUKvNe43n+GeAJbRJbpo11lccKVR1TdeyA==
            -----END AGE ENCRYPTED FILE-----
    lastmodified: "2023-08-27T19:02:02Z"
    mac: ENC[AES256_GCM,data:AgH0gLKrH537fjevKuiebwA9kw52Tl/TWSH6HctLcyiO8YHwsMJtQRZxHPKhJlNkFRV8BHjExfMgKEijOeMFfmOeF0yG73/3vk8o5hVtDIj2d8J4pLWvXeA7FU4/js8Z4u8JTvjewap4jHVMDlrlVGOTthJZ9ahW/hLF3EoCZtI=,iv:/rDngTcPDndx2UEj/5c3au1JPIiSWle2Rrd7F+KlbH0=,tag:9xI0inYNWbqTyA6ob61Qyg==,type:str]
    pgp: []
    encrypted_regex: ^(data|stringData)$
    version: 3.7.3
`

	want := `apiVersion: metal3.io/v1alpha1
kind: BareMetalHost
metadata:
  name: k8s-nodes-a-1
spec:
  automatedCleaning: disabled
  bmc:
    address: redfish-virtualmedia+http://host1/redfish/v1/Systems/7d7a911f-39db-478c-ba73-6e00bbdcf211
    credentialsName: k8s-nodes-a
  bootMACAddress: 52:54:00:6c:3c:01
  online: true
  rootDeviceHints:
    minSizeGigabytes: 10
---
apiVersion: metal3.io/v1alpha1
kind: BareMetalHost
metadata:
  name: k8s-nodes-a-2
spec:
  automatedCleaning: disabled
  bmc:
    address: redfish-virtualmedia+http://host2/redfish/v1/Systems/7d7a911f-39db-478c-ba73-6e00bbdcf212
    credentialsName: k8s-nodes-a
  bootMACAddress: 52:54:00:6c:3c:02
  online: true
  rootDeviceHints:
    minSizeGigabytes: 10
---
apiVersion: metal3.io/v1alpha1
kind: BareMetalHost
metadata:
  name: k8s-nodes-b-1
  namespace: test-ns
spec:
  automatedCleaning: disabled
  bmc:
    address: idrac-virtualmedia+http://host3:9000/redfish/v1/Systems/7d7a911f-39db-478c-ba73-6e00bbdcf213
    credentialsName: k8s-nodes-b
  bootMACAddress: 52:54:00:6c:3c:03
  online: true
  rootDeviceHints:
    minSizeGigabytes: 10
---
apiVersion: metal3.io/v1alpha1
kind: BareMetalHost
metadata:
  name: k8s-nodes-b-2
  namespace: test-ns
spec:
  automatedCleaning: disabled
  bmc:
    address: idrac-virtualmedia+http://host4:9000/redfish/v1/Systems/7d7a911f-39db-478c-ba73-6e00bbdcf214
    credentialsName: k8s-nodes-b
  bootMACAddress: 52:54:00:6c:3c:04
  online: true
  rootDeviceHints:
    minSizeGigabytes: 10
`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		if r.URL.Path == "/redfish/v1/" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"@odata.type": "#ServiceRoot.v1_5_0.ServiceRoot",
				"Id": "RedvirtService",
				"Name": "Redvirt Service",
				"RedfishVersion": "1.5.0",
				"UUID": "85775665-c110-4b85-8989-e6162170b3ec",
				"Systems": {
					"@odata.id": "/redfish/v1/Systems"
				},
				"@odata.id": "/redfish/v1/",
				"@Redfish.Copyright": "Copyright 2014-2016 Distributed Management Task Force, Inc. (DMTF). For the full DMTF copyright policy, see http://www.dmtf.org/about/policies/copyright."
			}`))
			return
		}
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		var manufacturer string
		switch r.URL.Hostname() {
		case "host1", "host2":
			manufacturer = "Unknown"
			if authHeader != "Basic Zm9vMTpiYXIx" {
				w.WriteHeader(http.StatusForbidden)
				return
			}
		case "host3", "host4":
			manufacturer = "Dell Inc."
			if authHeader != "Basic Zm9vMjpiYXIy" {
				w.WriteHeader(http.StatusForbidden)
				return
			}
		default:
			w.WriteHeader(http.StatusForbidden)
			return
		}

		hostIndex := strings.TrimPrefix(r.URL.Hostname(), "host")
		macAddr := fmt.Sprintf("52:54:00:6c:3c:0%s", hostIndex)
		systemId := fmt.Sprintf("7d7a911f-39db-478c-ba73-6e00bbdcf21%s", hostIndex)

		switch r.URL.Path {
		case "/redfish/v1/Systems":
			if err := template.Must(template.New("notImportant").Parse(`{
				"@odata.type": "#ComputerSystemCollection.ComputerSystemCollection",
				"Name": "Computer System Collection",
				"Members@odata.count": 1,
				"Members": [{
					"@odata.id": "/redfish/v1/Systems/{{ .system }}"
				}],
				"@odata.context": "/redfish/v1/$metadata#ComputerSystemCollection.ComputerSystemCollection",
				"@odata.id": "/redfish/v1/Systems",
				"@Redfish.Copyright": "Copyright 2014-2016 Distributed Management Task Force, Inc. (DMTF). For the full DMTF copyright policy, see http://www.dmtf.org/about/policies/copyright."
			}`)).Execute(w, map[string]interface{}{
				"system": systemId,
			}); err != nil {
				t.Fatal(err)
			}
		case fmt.Sprintf("/redfish/v1/Systems/%s", systemId):
			if err := template.Must(template.New("notImportant").Parse(`{
				"@odata.type": "#ComputerSystem.v1_1_0.ComputerSystem",
				"Id": "{{ .system }}",
				"Name": "k8s-node-1",
				"UUID": "{{ .system }}",
				"Manufacturer": "{{ .manufacturer }}",
				"EthernetInterfaces": {
					"@odata.id": "/redfish/v1/Systems/{{ .system }}/EthernetInterfaces"
				},
				"@odata.context": "/redfish/v1/$metadata#ComputerSystem.ComputerSystem",
				"@odata.id": "/redfish/v1/Systems/{{ .system }}",
				"@Redfish.Copyright": "Copyright 2014-2016 Distributed Management Task Force, Inc. (DMTF). For the full DMTF copyright policy, see http://www.dmtf.org/about/policies/copyright."
			}`)).Execute(w, map[string]interface{}{
				"system":       systemId,
				"manufacturer": manufacturer,
			}); err != nil {
				t.Fatal(err)
			}
		case fmt.Sprintf("/redfish/v1/Systems/%s", systemId):
			if err := template.Must(template.New("notImportant").Parse(`{
				"@odata.type": "#ComputerSystem.v1_1_0.ComputerSystem",
				"Id": "{{ .system }}",
				"Name": "k8s-node-1",
				"UUID": "{{ .system }}",
				"Manufacturer": "{{ .manufacturer }}",
				"EthernetInterfaces": {
					"@odata.id": "/redfish/v1/Systems/{{ .system }}/EthernetInterfaces"
				},
				"@odata.context": "/redfish/v1/$metadata#ComputerSystem.ComputerSystem",
				"@odata.id": "/redfish/v1/Systems/{{ .system }}",
				"@Redfish.Copyright": "Copyright 2014-2016 Distributed Management Task Force, Inc. (DMTF). For the full DMTF copyright policy, see http://www.dmtf.org/about/policies/copyright."
			}`)).Execute(w, map[string]interface{}{
				"system":       systemId,
				"manufacturer": manufacturer,
			}); err != nil {
				t.Fatal(err)
			}
		case fmt.Sprintf("/redfish/v1/Systems/%s/EthernetInterfaces", systemId):
			if err := template.Must(template.New("notImportant").Parse(`{
				"@odata.type": "#EthernetInterfaceCollection.EthernetInterfaceCollection",
				"Name": "Ethernet Interface Collection",
				"Description": "Virtual NICs",
				"Members@odata.count": 1,
				"Members": [{
					"@odata.id": "/redfish/v1/Systems/{{ .system }}/EthernetInterfaces/{{ .macAddr }}"
				}],
				"@odata.context": "/redfish/v1/$metadata#EthernetInterfaceCollection.EthernetInterfaceCollection",
				"@odata.id": "/redfish/v1/Systems/{{ .system }}/EthernetInterfaces"
			}`)).Execute(w, map[string]interface{}{
				"system":  systemId,
				"macAddr": macAddr,
			}); err != nil {
				t.Fatal(err)
			}
		case fmt.Sprintf("/redfish/v1/Systems/%s/EthernetInterfaces/%s", systemId, macAddr):
			if err := template.Must(template.New("notImportant").Parse(`{
				"@odata.type": "#EthernetInterface.v1_0_2.EthernetInterface",
				"Id": "{{ .macAddr }}",
				"Name": "VNIC {{ .macAddr }}",
				"Status": {
					"State": "Enabled",
					"Health": "OK"
				},
				"PermanentMACAddress": "{{ .macAddr }}",
				"MACAddress": "{{ .macAddr }}",
				"@odata.context": "/redfish/v1/$metadata#EthernetInterface.EthernetInterface",
				"@odata.id": "/redfish/v1/Systems/{{ .system }}/EthernetInterfaces/{{ .macAddr }}"
			}`)).Execute(w, map[string]interface{}{
				"system":  systemId,
				"macAddr": macAddr,
			}); err != nil {
				t.Fatal(err)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}))
	defer server.Close()
	serverUrl, err := url.Parse(server.URL)
	assert.NoError(t, err)
	httpClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(serverUrl)}}

	tempDir := t.TempDir()
	inputSrc := path.Join(tempDir, "secrets.yaml")
	err = os.WriteFile(inputSrc, []byte(input), 0600)
	assert.NoError(t, err)
	inputEncryptedSrc := path.Join(tempDir, "secrets-encrypted.yaml")
	err = os.WriteFile(inputEncryptedSrc, []byte(inputEncrypted), 0600)
	assert.NoError(t, err)

	outputDest := path.Join(tempDir, "hosts.yaml")
	// non-encrypted
	err = generateBareMetalHosts(inputSrc, outputDest, generateOptions{
		HTTPClient: httpClient,
	})
	assert.NoError(t, err)

	result, err := os.ReadFile(outputDest)
	assert.NoError(t, err)

	assert.Equal(t, want, string(result))

	// encrypted
	t.Setenv("SOPS_AGE_KEY", "AGE-SECRET-KEY-15JLZDHADZ45JVZXMKSAM9U8AHE47DDK7DTJL7XNR0G27U4P9XRHQLNKUH2")
	outputEncryptedDest := path.Join(tempDir, "hosts-encrypted.yaml")
	err = generateBareMetalHosts(inputEncryptedSrc, outputEncryptedDest, generateOptions{
		HTTPClient: httpClient,
	})
	assert.NoError(t, err)

	result, err = os.ReadFile(outputEncryptedDest)
	assert.NoError(t, err)

	assert.Equal(t, want, string(result))
}
