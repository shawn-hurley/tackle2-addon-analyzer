FROM registry.access.redhat.com/ubi9/ubi-minimal
USER root
WORKDIR /tmp
RUN microdnf -y update && microdnf -y clean all
RUN echo -e "[centos9]" \
 "\nname = centos9" \
 "\nbaseurl = http://mirror.stream.centos.org/9-stream/AppStream/\$basearch/os/" \
 "\nenabled = 1" \
 "\ngpgcheck = 0" > /etc/yum.repos.d/centos.repo
RUN microdnf -y install \
 java-17-openjdk-headless \
 openssh-clients \
 tar \
 wget \
 git \
 subversion \
 maven \
&& microdnf -y clean all
ENV HOME=/working \
 JAVA_HOME="/usr/lib/jvm/jre-17" \
 JAVA_VENDOR="openjdk" \
 JAVA_VERSION="17"
#
# Build jdtls.
#
ARG JDTLS=https://www.eclipse.org/downloads/download.php?file=/jdtls/milestones/1.6.0/jdt-language-server-1.6.0-202111261512.tar.gz
RUN wget -qO jdtls.tar.gz $JDTLS \
 && tar xvf jdtls.tar.gz -C /opt \
 && rm jdtls.tar.gz
#
# Build java provider bundle.
#
RUN mkdir -p m2 \
 && git clone https://github.com/konveyor/java-analyzer-bundle --depth 1 \
 && mvn clean install -Dmaven.repo.local=m2 -DskipTests=true -f java-analyzer-bundle/pom.xml \
 && cp m2/io/konveyor/tackle/java-analyzer-bundle.core/1.0.0-SNAPSHOT/java-analyzer-bundle.core-1.0.0-SNAPSHOT.jar \
 /opt
#
# Build analyzer and install gopls.
#
FROM registry.access.redhat.com/ubi9/go-toolset:latest as analyzer
RUN git clone https://github.com/konveyor/analyzer-lsp --depth 1
RUN mv analyzer-lsp/* .
ENV GOPATH=$APP_ROOT
RUN make build
RUN go install golang.org/x/tools/gopls@latest
#
# Build addon.
#
FROM registry.access.redhat.com/ubi9/go-toolset:latest as addon
ENV GOPATH=$APP_ROOT
COPY --chown=1001:0 . .
RUN make cmd
#
# Build container.
#
FROM registry.access.redhat.com/ubi9/ubi-minimal
WORKDIR /working
ARG GOPATH=/opt/app-root
COPY --from=addon $GOPATH/src/bin/addon /usr/local/bin
COPY --from=addon $GOPATH/src/provider-settings.json /working
COPY --from=analyzer $GOPATH/src/konveyor-analyzer /usr/local/bin
COPY --from=analyzer $GOPATH/src/konveyor-analyzer-dep /usr/local/bin
COPY --from=analyzer $GOPATH/bin/gopls /usr/local/bin
ENTRYPOINT ["/usr/local/bin/addon"]
