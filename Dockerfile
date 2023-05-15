FROM registry.access.redhat.com/ubi9/openjdk-17 as base
USER root
WORKDIR /tmp
RUN microdnf -y install \
 tar \
 wget \
 git
#
# Build jdtls.
#
ARG JDTLS=https://www.eclipse.org/downloads/download.php?file=/jdtls/milestones/1.16.0/jdt-language-server-1.16.0-202209291445.tar.gz
RUN wget -qO jdtls.tar.gz $JDTLS
RUN tar xvf jdtls.tar.gz -C /opt
#
# Build java provider bundle.
#
RUN mkdir -p m2
RUN git clone https://github.com/konveyor/java-analyzer-bundle --depth 1
RUN mvn clean install -Dmaven.repo.local=m2 -DskipTests=true -f java-analyzer-bundle/pom.xml
RUN cp m2/io/konveyor/tackle/java-analyzer-bundle.core/1.0.0-SNAPSHOT/java-analyzer-bundle.core-1.0.0-SNAPSHOT.jar \
 /opt/java-analyzer-bundle.jar
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
FROM registry.access.redhat.com/ubi9/openjdk-17
USER root
RUN echo -e "[centos9]" \
 "\nname = centos9" \
 "\nbaseurl = http://mirror.stream.centos.org/9-stream/AppStream/\$basearch/os/" \
 "\nenabled = 1" \
 "\ngpgcheck = 0" > /etc/yum.repos.d/centos.repo
RUN microdnf -y install \
 openssh-clients \
 subversion \
 git \
 tar
ENV HOME=/addon \
 ADDON=/addon \
 JAVA_HOME="/usr/lib/jvm/jre-17" \
 JAVA_VENDOR="openjdk" \
 JAVA_VERSION="17"
WORKDIR /addon
ARG GOPATH=/opt/app-root
COPY --from=base /opt ./opt
COPY --from=addon $GOPATH/src/bin/addon /usr/bin
COPY --from=addon $GOPATH/src/settings.yaml ./opt
COPY --from=analyzer $GOPATH/src/konveyor-analyzer ./opt
COPY --from=analyzer $GOPATH/src/konveyor-analyzer-dep ./opt
COPY --from=analyzer $GOPATH/bin/gopls ./opt
ENTRYPOINT ["/usr/bin/addon"]
