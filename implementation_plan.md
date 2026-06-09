# Implementation Plan - WebRTC Call & Disconnection Routing

This plan outlines the changes required to implement WebRTC call routing (`offer`, `answer`, `ice-candidate`, `hangup`) and robust peer disconnection handling.

## User Review Required

> [!IMPORTANT]
> The solution tracks peer connection state dynamically on the `Client` struct via a new `PeerID` field.
> - When a user starts or answers a call, the backend registers the partner's User ID (`PeerID`).
> - When a WebSocket connection disconnects abruptly, the backend automatically detects if they were in a call and dispatches a `"hangup"` event to the active call partner.

## Proposed Changes

### Backend WebRTC Call & Disconnection Flow

---

#### [MODIFY] [hub.go](file:///c:/Users/asus/Documents/GO_DEV/BLOG-WEB-GO-APP/Backend/services/hub.go)
- Add `PeerID uint` to the `Client` struct to track who the client is in a call with.
- Fix operator precedence and potential nil pointer dereference in `MessageReader` when parsing audio payloads.
- Extend `MessageReader` to parse and validate `"answer"` and `"hangup"` payloads.
- Update `PeerID` on the calling client when they send an `"offer"`, `"answer"`, or `"hangup"`.
- Modify `MessageReader`'s `defer` block to publish a `"hangup"` payload to the peer if `c.PeerID` is non-zero when the connection is severed.
- Refactor the `RunService`'s `AudioCallingOnly` handler:
  - Optimize client lookup by using `h.ClientStore[audioPayload.RecieverID]`.
  - Update `PeerID` on the destination client.
  - Forward `"offer"`, `"answer"`, `"ice-candidate"`, `"ice_candidate"`, and `"hangup"` payloads to the receiver.

#### [MODIFY] [pubsub_events.go](file:///c:/Users/asus/Documents/GO_DEV/BLOG-WEB-GO-APP/Backend/events/pubsub_events.go)
- Consolidate and expand the consumer message routing in `StartConsumingDeliveries()` to route all call-related payloads (`offer`, `answer`, `ice_candidate`, `ice-candidate`, `hangup`) to `p.hub.AudioCallingOnly`.

## Verification Plan

### Automated Tests
- Run `go build ./...` to verify compilation.
- We can add or run existing WebSocket tests if available.

### Manual Verification
- Verify call routing and abrupt disconnection using the frontend debugger or two simultaneous active connections.
