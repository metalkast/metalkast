import { systemManifest } from "../../../../../docs/utils/manifests";

async function main() {
    console.log(systemManifest({
        manifestsRef: "main",
        ingressIP: "192.168.123.105",
        ingressDomain: "192.168.123.105.nip.io",
        extraSystemConfigProperties: new Map([
            ["ingress_cert_email", "letsencrypt@metalkast.io"],
        ]),
        extraComponents: [
            "https://github.com/metalkast/metalkast//manifests/system/base/nginx-ingress/components/issuers/letsencrypt-cloudflare",
        ]
    }))
}

main()
