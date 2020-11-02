MACOS_BUILD_DIR = .macos/build
MACOS_APP_OUTPUT = out/prmon.app
OUTPUT_DIR = out

tui:
	mkdir -p ${OUTPUT_DIR}
	go build -o ${OUTPUT_DIR}/tui ./cmd/tui

# ! todo - add 2goarray as a build dep
# 1. generate-source-icons.sh - build the byte array representation for png icons
# 2. generate-iconset.sh - build the icons for the app bundle
icons:
	sh ${MACOS_BUILD_DIR}/generate-source-icons.sh CircleRegular circle_regular
	sh ${MACOS_BUILD_DIR}/generate-source-icons.sh CircleSolid circle_solid
	sh ${MACOS_BUILD_DIR}/generate-iconset.sh .macos/build/icon-1024.png

mac:
	mkdir -p ${OUTPUT_DIR}
	cp -r .macos/prmon.app ${MACOS_APP_OUTPUT}
	go build -o ${MACOS_APP_OUTPUT}/Contents/MacOS/prmon ./cmd/mac

clean:
	rm -rf ${OUTPUT_DIR}
