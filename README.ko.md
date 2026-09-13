# pixelduet

[English](README.md)

모델과 사람이 함께 쓰는 픽셀 캔버스입니다. 모델이 브라우저에 한 획씩 그리는 동안 지켜보다가, 멈추고 붓을 가져가 고치고 메모를 남긴 뒤 붓을 돌려줍니다. 모델은 다음 획의 응답으로 당신이 한 일을 알게 됩니다.

![모델이 개구리를 그리고, 사람이 볼을 칠하고, 모델이 왕관을 얹는 장면](docs/demo.gif)

## 설치

    npx -y pixelduet version

또는

    go install github.com/kimjungminn24/pixel-duet@latest

또는 Releases 페이지에서 바이너리를 내려받습니다. Linux, macOS, Windows를 지원합니다.

## 빠른 시작

    npx -y pixelduet

캔버스 서버와 브라우저 뷰어를 띄우고 뷰어를 엽니다.

Claude Code는 한 번만 연결합니다.

    claude mcp add pixelduet -- npx -y pixelduet bridge

세션 시작 시 실행 중인 서버가 없으면 브리지가 직접 띄웁니다.

이 저장소를 체크아웃했다면 `go run .`이 `npx -y pixelduet`을 대신하고, Claude Code는 `.mcp.json`에서 브리지를 읽습니다.

## 다른 MCP 클라이언트

브리지는 stdio 위의 MCP 서버입니다. Codex CLI:

    codex mcp add pixelduet -- npx -y pixelduet bridge

또는 `~/.codex/config.toml`에:

    [mcp_servers.pixelduet]
    command = "npx"
    args = ["-y", "pixelduet", "bridge"]

Windows에서는 `command = "cmd"`, `args = ["/c", "npx", "-y", "pixelduet", "bridge"]`로 적습니다.

새 클라이언트에서 확인할 두 가지:

- 그리기 규칙은 서버의 `instructions`로 전달됩니다. 클라이언트가 이를 버리면 `npx -y pixelduet rules >> AGENTS.md`로 되살립니다.
- `get_canvas`는 PNG를 돌려줍니다. 이미지를 전달하지 않는 클라이언트에서는 모델이 `read_pixels` 숫자만 보게 됩니다.

MCP 없이도 TCP 소켓을 열 수 있는 것이면 무엇이든 그릴 수 있습니다. 프로토콜은 internal/proto/proto.go에 있습니다.

## 붓 넘기기

1. 모델이 호출 한 번에 한 획씩 그리며 로그에 설명을 남깁니다.
2. 일시정지를 누릅니다. 모델의 다음 획은 붙잡힙니다.
3. 픽셀을 고치고, 색을 고르고, 메모를 남깁니다. 당신의 편집은 기다리지 않습니다.
4. 재개를 누릅니다. 붙잡혔던 획이 그려지고 응답에 이렇게 적힙니다.

       ok changed=1 waited=12s edits=5 box=9,8-13,10 note=make the eyes bigger

5. 모델이 캔버스를 읽고, 당신의 편집을 그대로 두고, 계속 그립니다.

## 말 걸기

로그 아래 입력창에 언제든 적습니다.

- 모델이 그리는 중이면 다음 획의 응답에 당신의 말이 실립니다.
- 그릴 것이 없으면 모델은 `listen`을 부르고, 당신이 말하면 돌아옵니다.

메모는 순서대로 쌓입니다. 모델은 같은 로그에 `say`로 답합니다.

## 뷰어

| | |
|---|---|
| 도구 | 펜 B, 지우개 E, 직선 L, 사각형 R, 원 O, 채운 사각형 U, 채운 원 P, 채우기 F, 추출 I |
| 캔버스 | 배율 슬라이더, 맞춤, 이동 H 또는 Space+드래그, 격자 G, 되돌리기 Ctrl+Z, 크기 32/64/128 |
| 팔레트 | DB32 32칸. 한 칸을 고치면 그 칸을 쓰는 픽셀이 모두 바뀝니다 |
| 모델 | 재생/정지, 획 사이에 최대 3초를 더하는 속도 슬라이더 |
| 저장 | 서버의 갤러리 폴더, 또는 직접 고른 폴더(Chrome, Edge) |
| 언어 | 브라우저 언어를 따라 한국어 또는 영어. 헤더 버튼으로 전환, `?lang=en`으로 고정 |

## 모델이 받는 것

- `get_canvas`: 캔버스 전체 또는 일부를 PNG로, 거울상과 다른 픽셀 수와 함께.
- `read_pixels`: 영역을 숫자로, 더 짧으면 런렝스 행으로.
- `draw_pixels`, `draw_line`, `draw_rect`, `draw_ellipse`, `fill_area`. 모두 `mirror`로 중심선 반대편에 반사를 그립니다.
- `shade`, `highlight`, `outline`, `dither`: 이미 칠한 것에 대한 마무리 작업.
- `list_sprites`, `view_sprite`, `save_sprite`, `stamp`: `sprites/`에 저장된 그리드 갤러리.
- `say`, `listen`: 채팅.
- 접속 시 전달되는 그리기 규칙: 비율을 숫자로 계획, 실루엣부터, 단계마다 그림 확인, 광원은 하나, 넘어가기 전에 묻기. internal/bridge/rules.go 참고.

## 구조

    pixelduet serve            캔버스를 소유
      |- pixelduet web         브라우저 <-> HTTP/SSE <-> 줄 프로토콜
      `- pixelduet bridge      MCP 클라이언트 <-> 도구 <-> 줄 프로토콜

`pixelduet`만 실행하면 serve와 web이 한 프로세스에서 돕니다. 서버가 없으면 브리지도 같은 일을 합니다.

캔버스는 팔레트 인덱스의 격자로, 픽셀당 한 글자(`0` 빈칸, `1`~`w` 색)이며 전송, 저장 파일, 모델이 읽는 형식이 같습니다. 뷰어는 internal/web/static/ 아래의 HTML, CSS, ES 모듈이고 바이너리에 내장됩니다.

## 테스트

    go test ./...
    node --test test/*.test.mjs

[CONTRIBUTING.md](CONTRIBUTING.md)를 참고하세요.

## 라이선스

MIT. [LICENSE](LICENSE) 참고.
