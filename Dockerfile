FROM runnergo/debian:stable-slim

ADD  manage  /data/manage/manage


CMD ["/data/manage/manage","-m", "1"]
