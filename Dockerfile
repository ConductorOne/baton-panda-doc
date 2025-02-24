FROM gcr.io/distroless/static-debian11:nonroot
ENTRYPOINT ["/baton-panda-doc"]
COPY baton-panda-doc /