FROM ubuntu:22.04
ARG KUBECTL_VERSION=v1.27.1

RUN apt-get update -y && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y \
        build-essential \
        qemu-kvm \
        libvirt-daemon-system \
        libvirt-clients \
        bridge-utils \
        jq \
        curl \
        supervisor \
        virtinst \
        libnss-libvirt \
        python3 \
        python3-dev \
        python3-pip \
        openssh-client \
        ca-certificates \
    	git \
    	zsh \
    	vim \
    	dnsutils \
        nginx \
        fcgiwrap \
        spawn-fcgi \
        # ipmi console access
        ipmitool \
        freeipmi-tools \
        # to expect output from ipmi console
        expect \
        # UEFI support
        ovmf \
        && \
    apt-get autoclean && \
    apt-get autoremove

RUN pip install sushy-tools libvirt-python git+https://github.com/metalkast/virtualbmc.git

RUN curl -L https://download.stateful.com/runme/1.7.3/runme_linux_x86_64.deb -o runme_linux_x86_64.deb && \
    dpkg -i runme_linux_x86_64.deb && \
    rm runme_linux_x86_64.deb

RUN mkdir -p /root/zsh-completions
RUN echo 'export fpath=(/root/zsh-completions $fpath)' >> /root/.zshrc
RUN echo "rm -f ~/.zcompdump; compinit" >> /root/.zshrc
RUN echo "alias k=kubectl" >> /root/.zshrc
RUN chsh -s $(which zsh)

RUN curl -L https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl \
    -o /usr/local/bin/kubectl && \
    chmod +x /usr/local/bin/kubectl && \
    kubectl completion zsh > /root/zsh-completions/_kubectl

RUN <<EOF
(
  set -x; cd "$(mktemp -d)" &&
  OS="$(uname | tr '[:upper:]' '[:lower:]')" &&
  ARCH="$(uname -m | sed -e 's/x86_64/amd64/' -e 's/\(arm\)\(64\)\?.*/\1\2/' -e 's/aarch64$/arm64/')" &&
  KREW="krew-${OS}_${ARCH}" &&
  curl -fsSLO "https://github.com/kubernetes-sigs/krew/releases/latest/download/${KREW}.tar.gz" &&
  tar zxvf "${KREW}.tar.gz" &&
  ./"${KREW}" install krew
)
echo 'export PATH="${KREW_ROOT:-$HOME/.krew}/bin:$PATH"' >> /root/.zshrc
zsh -c 'source /root/.zshrc; kubectl krew install ctx ns konfig'
EOF

COPY kube-ps1.patch /
RUN <<EOF
git clone https://github.com/jonmosco/kube-ps1.git /opt/kube-ps1
(
    cd /opt/kube-ps1; git apply /kube-ps1.patch; rm -f /kube-ps1.patch
)
echo "
source /opt/kube-ps1/kube-ps1.sh
PROMPT='\$(kube_ps1)'\$PROMPT" >> /root/.zshrc
EOF

RUN <<EOF
curl https://baltocdn.com/helm/signing.asc | gpg --dearmor | tee /usr/share/keyrings/helm.gpg > /dev/null
apt-get install apt-transport-https --yes
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/helm.gpg] https://baltocdn.com/helm/stable/debian/ all main" | tee /etc/apt/sources.list.d/helm-stable-debian.list
apt-get update
apt-get install helm
EOF

RUN mkdir -p /root/.kube

RUN sed -i -E 's/(hosts:\s+files)/\1 libvirt_guest/' /etc/nsswitch.conf

RUN sed -i 's/listen 80/listen 192.168.122.1:80/' /etc/nginx/sites-enabled/default
COPY nginx-git /etc/nginx/sites-enabled/

RUN mkdir -p /opt/config
COPY lab-network.xml /opt/config/lab-network.xml
COPY supervisord.conf /etc/supervisord.conf
COPY hosts-setup.sh /usr/bin/hosts-setup.sh
COPY redfish.sh /usr/bin/redfish.sh
COPY dummy-host.sh /usr/bin/dummy-host.sh
COPY copy-kubeconfigs.sh /usr/bin/copy-kubeconfigs.sh
COPY setup-workspace-git-repo.sh /usr/bin/setup-workspace-git-repo.sh
COPY htpasswd /opt/config/htpasswd
COPY sushy.conf.tmpl /opt/config/sushy.conf.tmpl

COPY entrypoint.sh /entrypoint.sh
CMD ["/entrypoint.sh"]
