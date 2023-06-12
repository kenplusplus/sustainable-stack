FROM quay.io/centos/centos:stream8

RUN yum install -y epel-release epel-next-release && \
    yum install cpuid-20230505-2.el8.x86_64 -y

ADD platform-validation.sh /usr/bin/platform-validation.sh

ENTRYPOINT ["/usr/bin/platform-validation.sh"]
