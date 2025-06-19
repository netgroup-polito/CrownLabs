FROM alpine:3.12.1
RUN apk add --no-cache dumb-init openssh

# Create new user bastion with nologin
RUN adduser -D -s /sbin/nologin bastion
RUN passwd -u -d bastion
RUN mkdir /ssh_pids && chmod 777 /ssh_pids

# sshd configuration file
COPY ./sshd_config_custom /etc/ssh/sshd_config_custom

# welcome message to be displayed in case the user does not use the -J option
COPY ./motd /etc/motd

EXPOSE 2222

ENTRYPOINT ["/usr/bin/dumb-init", "--"]

CMD ["/usr/sbin/sshd", "-D", "-e", "-f", "/etc/ssh/sshd_config_custom"]