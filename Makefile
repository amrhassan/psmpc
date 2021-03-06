GO=go
EXECNAME=psmpc
prefix := /usr/local

${EXECNAME}: *.go */*.go
	${GO} build -o "${EXECNAME}"

install: ${EXECNAME}
	install -Dm755 "${EXECNAME}" "${prefix}/bin/${EXECNAME}"
	install -Dm755 "gui/ui.glade" "${prefix}/share/${EXECNAME}/gui/ui.glade"
	install -Dm755 "gui/icon.png" "${prefix}/share/${EXECNAME}/gui/icon.png"
	install -Dm755 "gui/album.png" "${prefix}/share/${EXECNAME}/gui/album.png"

uninstall:
	rm -rf "$(prefix)/bin/$(EXECNAME)" "$(prefix)/share/$(EXECNAME)"
