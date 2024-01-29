FROM alpine:3.19.0

RUN apk add \
  build-base \
  ipmitool \
  python3-dev \
  py3-alembic \
  py3-bcrypt \
  py3-cryptography \
  py3-greenlet \
  py3-lxml \
  py3-mako \
  py3-markupsafe \
  py3-netifaces \
  py3-pip \
  py3-prettytable \
  py3-psutil \
  py3-pyrsistent \
  py3-pysendfile \
  py3-simplejson \
  py3-sqlalchemy \
  py3-tempita \
  py3-yaml \
  py3-wheel \
  py3-wrapt \
  sqlite \
  xorriso \
  qemu-img \
  syslinux

ARG IRONIC_VERSION=23.1.0

RUN pip3 install --break-system-packages --no-cache-dir \
  ironic==${IRONIC_VERSION} \
  ironic-inspector \
  python-dracclient \
  python-ibmcclient \
  sushy \
  sushy-oem-idrac

COPY ironic.conf ironic-inspector.conf etc/defaults/

RUN mkdir -p /var/lib/ironic-inspector && sqlite3 /var/lib/ironic-inspector/ironic-inspector.db "pragma journal_mode=wal"
