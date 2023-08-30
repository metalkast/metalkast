# TODO (@danilo404): Fix pip versions after testing, move the image to a proper place
# TODO (GAL-251): Decide how to handle the ironic image and configuration, templating, etc.
FROM alpine:3.17.2

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

RUN pip3 install --no-cache-dir \
  ironic \
  ironic-inspector \
  python-dracclient \
  python-ibmcclient \
  sushy \
  sushy-oem-idrac \
  # Flask 2.3 removes some API and there isn't a fix for this in ironic
  Flask==2.2.3

COPY ironic.conf ironic-inspector.conf etc/defaults/

RUN mkdir -p /var/lib/ironic-inspector && sqlite3 /var/lib/ironic-inspector/ironic-inspector.db "pragma journal_mode=wal"
