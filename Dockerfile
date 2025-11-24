FROM scratch
COPY --chmod=755 main ./main
COPY version src/index.tmpl ./
ENTRYPOINT [ "./main" ]