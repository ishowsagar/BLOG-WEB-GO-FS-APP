import React, {
  useEffect,
  useRef,
  useState,
  useImperativeHandle,
  forwardRef,
} from "react";
import { useOutletContext } from "react-router-dom";

const AudioCalling = forwardRef(
  (
    {
      passedCurrentUserID,
      subscribeNotifications: propSubscribeNotifications,
      sendNotifications: propSendNotifications,
    },
    ref,
  ) => {
    // states
    const [callState, setCallState] = useState("idle"); //["idle","calling","hangup","active"."incoming"]
    const [incomingCallFrom, setIncomingCallFrom] = useState(null); //* for tracking who is calling
    const [incomingOfferSdp, setIncomingOfferSdp] = useState(null);
    const [seconds, setSeconds] = useState(0);
    const [finalDuration, setFinalDuration] = useState("");

    console.log("/audio_calling state:", callState);

    // webRtc & network refs
    const peerConnection = useRef(null);
    const localStream = useRef(null);
    const targetPeerID = useRef(null); //* the one current user is connected to <- sent by the backend answer
    const iceCandidatesQueue = useRef([]);

    // stun and turn servers configuration
    const rtcConfig = {
      iceTransportPolicy: "all",
      iceServers: [
        // One robust master STUN server handles 99% of P2P lookups
        { urls: "stun:stun.l.google.com:19302" },

        // Streamlined TURN relay channels favoring stable ports
        {
          urls: "turn:openrelay.metered.ca:80",
          username: "openrelayproject",
          credential: "openrelayproject",
        },
        {
          urls: "turn:openrelay.metered.ca:443",
          username: "openrelayproject",
          credential: "openrelayproject",
        },
        {
          urls: "turn:openrelay.metered.ca:443?transport=tcp", // Forces reliable TCP over flaky UDP
          username: "openrelayproject",
          credential: "openrelayproject",
        },
      ],
    };

    const outletContext = useOutletContext();
    const sendNotifications =
      propSendNotifications || outletContext?.sendNotifications;
    const subscribeNotifications =
      propSubscribeNotifications || outletContext?.subscribeNotifications;

    const logTrackDetails = (prefix, track) => {
      if (!track) {
        console.log(`[WebRTC Telemetry] ${prefix} - No track provided`);
        return;
      }
      try {
        const settings = track.getSettings ? track.getSettings() : {};
        const capabilities = track.getCapabilities ? track.getCapabilities() : {};
        console.log(`[WebRTC Telemetry] ${prefix} Track Details:`, {
          id: track.id,
          kind: track.kind,
          label: track.label,
          enabled: track.enabled,
          muted: track.muted,
          readyState: track.readyState,
          settings: JSON.stringify(settings),
          capabilities: JSON.stringify(capabilities),
        });
        // ! note: do NOT set onmute/onunmute here - caller sets them in ontrack handler
        // ! overwriting them here would kill the audio restart logic
        track.onended = () => {
          console.log(`[WebRTC Telemetry] ${prefix} Track ENDED:`, track.id);
        };
      } catch (err) {
        console.error(`[WebRTC Telemetry] Error logging track details for ${prefix}:`, err);
      }
    };

    const setupPeerConnectionListeners = (pc) => {
      console.log("[WebRTC Telemetry] Setting up listeners on RTCPeerConnection");

      pc.oniceconnectionstatechange = () => {
        const state = pc.iceConnectionState;
        console.log(`[WebRTC Telemetry] ICE Connection State changed to: ${state}`);
        if (state === "failed" || state === "disconnected") {
          console.warn("[WebRTC Telemetry] ICE Connection is in failed/disconnected state.");
        }
        // ! belt-and-suspenders: when ICE confirms a live path, re-trigger audio element play()
        // ! in case ontrack fired before ICE and the initial play() hit silence
        if (state === "connected" || state === "completed") {
          const el = document.getElementById("remote-hidden-audio-element");
          if (el && el.srcObject) {
            console.log("[WebRTC Telemetry] ICE connected - re-triggering audio element play()");
            el.muted = false;
            el.volume = 1.0;
            el.play().catch((err) => console.error("[WebRTC Telemetry] Play on ICE connect failed:", err));
          }
        }
      };

      pc.onsignalingstatechange = () => {
        console.log(`[WebRTC Telemetry] Signaling State changed to: ${pc.signalingState}`);
      };

      pc.onconnectionstatechange = () => {
        console.log(`[WebRTC Telemetry] Connection State changed to: ${pc.connectionState}`);
      };

      pc.onicegatheringstatechange = () => {
        console.log(`[WebRTC Telemetry] ICE Gathering State changed to: ${pc.iceGatheringState}`);
      };

      pc.onicecandidateerror = (e) => {
        console.error("[WebRTC Telemetry] ICE Candidate Error:", {
          errorCode: e.errorCode,
          errorText: e.errorText,
          url: e.url
        });
      };
    };

    // Cleanup/Hangup handling
    const handleHangup = () => {
      // ! whem hanged up safely cleam dead rtc connection or open streams n mics
      if (peerConnection.current) {
        peerConnection.current.ontrack = null;
        peerConnection.current.onremovetrack = null;
        peerConnection.current.onicecandidate = null;
        peerConnection.current.oniceconnectionstatechange = null;
        peerConnection.current.onsignalingstatechange = null;

        // closing the peer rtc connection
        peerConnection.current.close();
        peerConnection.current = null;
      }

      // since local stream carries the current caller media devices
      if (localStream.current) {
        localStream.current?.getTracks().forEach((track) => {
          track.stop(); // stopping all mic streams
        });
        localStream.current = null;
      }


      // tearing down the hidden audio served element for tracks
      const remoteAudioElement = document.getElementById(
        "remote-hidden-audio-element",
      );
      if (remoteAudioElement) {
        remoteAudioElement.srcObject = null;
      }

      iceCandidatesQueue.current = [];
      setCallState("hangup");
      console.log(
        "user has been disconnected from the call; all the rtc connection have been safely remved.",
      );
    };

    const triggerHangup = (notifyPeer = true) => {
      if (notifyPeer && sendNotifications) {
        const peerId = targetPeerID.current || incomingCallFrom;
        if (peerId) {
          sendNotifications({
            sender_id: Number(passedCurrentUserID),
            reciever_id: Number(peerId),
            type: "hangup",
          });
        }
      }
      handleHangup();
    };

    // Expose IntialiseCalling trigger to the parent component
    useImperativeHandle(ref, () => ({
      initialiseCalling: (recieverID) => {
        IntialiseCalling(recieverID);
      },
    }));

    // Incoming call accept handler
    const acceptCall = async () => {
      try {
        console.log("[WebRTC Accept] Accepting incoming call from:", incomingCallFrom);
        if (!incomingOfferSdp) {
          console.warn("[WebRTC Accept] No incomingOfferSdp available!");
          return;
        }

        // ! pre-activate the remote audio element with the user click gesture to bypass autoplay policy blocks
        const el = document.getElementById("remote-hidden-audio-element");
        if (el) {
          el.muted = false;
          el.volume = 1.0;
          el.play().catch((err) => console.log("[WebRTC Accept] Pre-play gesture activation:", err.message));
        }

        // peer connection is stored in current state <- created by new RTCPeerConnection(passingInIceStunServersConfig)
        peerConnection.current = new RTCPeerConnection(rtcConfig); // just remember everything is stored in current state by the use of ref
        setupPeerConnectionListeners(peerConnection.current);

        // before setting up connection,adding this to candidate for ipLookup by stun servers and candidate path matching for interconnection
        peerConnection.current.onicecandidate = (e) => {
          //**  sending payload of audio_type payload of context - e.candidate => if e.candidate exists in peerConn. here --
          if (e.candidate) {
            console.log("[WebRTC Accept] Generated local ICE Candidate (Acceptor):", {
              candidate: e.candidate.candidate,
              sdpMid: e.candidate.sdpMid,
              sdpMLineIndex: e.candidate.sdpMLineIndex
            });
            if (sendNotifications) {
              sendNotifications({
                sender_id: Number(passedCurrentUserID),
                reciever_id: Number(incomingCallFrom),
                type: "ice-candidate",
                audio_payload_only: e.candidate,
              });
            }
          } else {
            console.log("[WebRTC Accept] ICE Candidate Gathering complete (Acceptor)");
          }
        };

        // streaming incoming stream from peerRtcConnection
        peerConnection.current.ontrack = (e) => {
          const track = e.track; // always use the track directly, not e.streams[0] which can be muted/empty on first fire
          // on track property gives us remotely recieved stream <- in streams array being the 0th as first element being the recieved stream
          const remoteAudioStream = e.streams[0] || new MediaStream([track]); // fallback: build a stream from the track itself if streams[0] is missing

          console.log("[WebRTC Accept] ontrack event received:", {
            streamsCount: e.streams.length,
            trackKind: track.kind,
            trackId: track.id,
            trackMuted: track.muted,
            trackReadyState: track.readyState
          });
          logTrackDetails("Remote (Accept)", track);

          let remoteHiddenAudioElement = document.getElementById(
            "remote-hidden-audio-element",
          );

          // if (!remoteHiddenAudioElement) {
          //   console.log("[WebRTC Accept] Creating remote-hidden-audio-element");
          //   remoteHiddenAudioElement = document.createElement("audio");
          //   remoteHiddenAudioElement.id = "remote-hidden-audio-element";
          //   document.body.appendChild(remoteHiddenAudioElement);
          // }

          // if element exists and cause it would be hidden, sourcing the recieved stream from peerConnection rtc to source in from e.streams[at0thPlace]
          if (remoteHiddenAudioElement) {
            // ! setting {src} of this hidden el to play this audio track steam from the candidate
            if (remoteHiddenAudioElement.srcObject !== remoteAudioStream) {
              console.log("[WebRTC Accept] Setting remoteHiddenAudioElement.srcObject to remoteAudioStream");
              remoteHiddenAudioElement.srcObject = remoteAudioStream;
            }
            remoteHiddenAudioElement.autoplay = true;
            remoteHiddenAudioElement.playsInline = true;
            remoteHiddenAudioElement.muted = false; // explicitly unmute - browser can default to muted for autoplay policy

            // ! critical: ontrack fires BEFORE ICE connects - track starts muted, play() hits silence.
            // ! bind onunmute to restart playback the moment ICE connects and audio starts flowing
            track.onunmute = () => {
              console.log("[WebRTC Accept] Track UNMUTED (ICE connected, audio flowing) - restarting play()");
              remoteHiddenAudioElement.play().catch((err) => {
                console.error("[WebRTC Accept] Audio playback failed on unmute:", err);
              });
            };

            console.log("[WebRTC Accept] Playing remote audio...");
            remoteHiddenAudioElement.play()
              .then(() => {
                console.log("[WebRTC Accept] Playback started successfully.");
              })
              .catch((err) => {
                console.error("[WebRTC Accept] Audio playback failed:", err);
              });
          }
        };

        // setting session description to be from recieved 'offer' payload from ws connection
        console.log("[WebRTC Accept] Setting remote description (Offer SDP)...");
        await peerConnection.current.setRemoteDescription(
          new RTCSessionDescription(incomingOfferSdp),
        );
        console.log("[WebRTC Accept] Set remote description SUCCESS");

        // Process queued candidates
        if (iceCandidatesQueue.current.length > 0) {
          console.log(
            `[WebRTC Accept] Processing ${iceCandidatesQueue.current.length} queued ICE candidates on accept`,
          );
          for (const candidate of iceCandidatesQueue.current) {
            // ! skip null/end-of-gathering candidates
            if (!candidate || (!candidate.candidate && candidate.sdpMid == null && candidate.sdpMLineIndex == null)) {
              console.log("[WebRTC Accept] Skipping null/end-of-gathering queued ICE candidate");
              continue;
            }
            try {
              console.log("[WebRTC Accept] Adding queued candidate:", candidate);
              await peerConnection.current.addIceCandidate(
                new RTCIceCandidate(candidate),
              );
              console.log("[WebRTC Accept] Added queued candidate SUCCESS");
            } catch (err) {
              console.error("[WebRTC Accept] Failed to add queued ICE candidate:", err);
            }
          }
          iceCandidatesQueue.current = [];
        }

        // get his mics and all and store in localStream as streamingMic
        console.log("[WebRTC Accept] Requesting local microphone stream...");
        localStream.current = await window.navigator.mediaDevices.getUserMedia({
          audio: true,
          video: false,
        });
        console.log("[WebRTC Accept] Local microphone acquired, stream ID:", localStream.current?.id);
        localStream.current.getTracks().forEach((track, i) => {
          logTrackDetails(`Local [${i}]`, track);
        });

        // add local tracks
        localStream.current.getTracks().forEach((track) => {
          console.log("[WebRTC Accept] Adding local track to connection:", track.id);
          peerConnection.current.addTrack(track, localStream.current);
        });

        // now sending answer payload to the caller with sdp block and audio_payload, this time sending 'answer' payload,
        console.log("[WebRTC Accept] Creating answer...");
        const createdAnsSdpPayload =
          await peerConnection.current.createAnswer({ offerToReceiveAudio: true }); //sdp answer audio payload, sending to caller with type "answer" for connection
        console.log("[WebRTC Accept] Create answer SUCCESS. Setting local description...");
        await peerConnection.current.setLocalDescription(createdAnsSdpPayload);
        console.log("[WebRTC Accept] Set local description SUCCESS");

        // * every offer/ans/candidate_req payload is sent to ws, attaching audio_payload based off context of outbound requests
        if (sendNotifications) {
          console.log("[WebRTC Accept] Sending answer notification via WS to targetPeerID:", targetPeerID.current);
          sendNotifications({
            sender_id: Number(passedCurrentUserID),
            reciever_id: Number(targetPeerID.current), // senderId - id of the user whose 'offer' request was intercepted and in response 'answer' is created for him wiht ans audio_payload attached inside
            type: "answer", // sending ans payload to the ws with type being "answer" -> by unmarshaling it would know what is incoming audio_payload and what needs to be published & shipped to consume by the reciever
            audio_payload_only: createdAnsSdpPayload,
          });
        } else {
          console.warn("[WebRTC Accept] sendNotifications is not available to send ANSWER");
        }

        setCallState("active");
      } catch (error) {
        console.error("[WebRTC Accept] Failed to accept incoming call:", error);
        triggerHangup(true);
      }
    };

    // ws connection is already there, just subcribing to it and retreving information
    useEffect(() => {
      if (!subscribeNotifications) return;

      const unsubscribe = subscribeNotifications(async (audioPayload) => {
        //** interceptor recieved the payload on onmessage
        if (!audioPayload) return;

        switch (
        audioPayload.type //& when call 'offer' is recieved <- for cal connection => reciever get that request with peerID being the senderID as sender is one who is sending call request
        ) {
          case "offer": {
            console.log(`[WebRTC WS] Received OFFER from sender: ${audioPayload.sender_id}`, {
              type: audioPayload.type,
              sdpLength: audioPayload.audio_payload_only?.sdp?.length || 0,
              sdpType: audioPayload.audio_payload_only?.type
            });
            targetPeerID.current = audioPayload.sender_id; // sender_id is what supplied by hub from reciever attaching -> the senderID ;reciever sends the sender id as peerID
            setIncomingCallFrom(audioPayload.sender_id);
            setIncomingOfferSdp(audioPayload.audio_payload_only);
            setCallState("incoming");
            break;
          }
          case "answer": {
            console.log(`[WebRTC WS] Received ANSWER from sender: ${audioPayload.sender_id}`, {
              type: audioPayload.type,
              sdpLength: audioPayload.audio_payload_only?.sdp?.length || 0,
              sdpType: audioPayload.audio_payload_only?.type
            });
            // &when call is either approved or not <- 'answering' call => reciever gets 'answer''s audio_payload
            if (!peerConnection.current) {
              console.warn("[WebRTC WS] Received ANSWER but peerConnection.current is null!");
              break;
            }
            // ! guard: only set remote description if we're in the right signaling state
            // ! if 3 handlers fire for this same message, 2nd and 3rd would throw "wrong state" without this check
            if (peerConnection.current.signalingState !== "have-local-offer") {
              console.warn(`[WebRTC WS] Skipping setRemoteDescription(answer) - wrong signalingState: ${peerConnection.current.signalingState}`);
              break;
            }
            try {
              console.log("[WebRTC WS] Setting remote description (Answer)...");
              await peerConnection.current.setRemoteDescription(
                new RTCSessionDescription(audioPayload.audio_payload_only), // opening rtcConnection from recieved audioPayload to the reciever
              );
              console.log("[WebRTC WS] Set remote description SUCCESS");
              setCallState("active");

              // Process queued candidates
              if (iceCandidatesQueue.current.length > 0) {
                console.log(
                  `[WebRTC WS] Processing ${iceCandidatesQueue.current.length} queued ICE candidates on answer`,
                );
                for (const candidate of iceCandidatesQueue.current) {
                  // ! skip null/end-of-gathering candidates
                  if (!candidate || (!candidate.candidate && candidate.sdpMid == null && candidate.sdpMLineIndex == null)) {
                    console.log("[WebRTC WS] Skipping null/end-of-gathering queued ICE candidate");
                    continue;
                  }
                  try {
                    console.log("[WebRTC WS] Adding queued ICE candidate:", candidate);
                    await peerConnection.current.addIceCandidate(
                      new RTCIceCandidate(candidate),
                    );
                    console.log("[WebRTC WS] Added queued ICE candidate SUCCESS");
                  } catch (err) {
                    console.error("[WebRTC WS] Failed to add queued ICE candidate:", err);
                  }
                }
                iceCandidatesQueue.current = [];
              }
            } catch (err) {
              console.error("[WebRTC WS] Failed to process ANSWER:", err);
            }
            break;
          }
          case "ice-candidate": {
            //& when both gets connected
            const candidate = audioPayload.audio_payload_only;
            console.log(`[WebRTC WS] Received ICE-CANDIDATE from sender: ${audioPayload.sender_id}`, candidate);

            // ! guard: null candidate = end-of-gathering signal from the browser, not a real candidate
            // ! constructing RTCIceCandidate from it throws "sdpMid and sdpMLineIndex are both null"
            if (!candidate || (!candidate.candidate && candidate.sdpMid == null && candidate.sdpMLineIndex == null)) {
              console.log("[WebRTC WS] Skipping null/end-of-gathering ICE candidate");
              break;
            }

            if (
              peerConnection.current &&
              peerConnection.current.remoteDescription // since we store connection when answered <- if that exists
            ) {
              try {
                console.log("[WebRTC WS] Adding ICE candidate directly to peerConnection...");
                await peerConnection.current.addIceCandidate(
                  new RTCIceCandidate(candidate),
                );
                console.log("[WebRTC WS] Added ICE candidate SUCCESS");
              } catch (err) {
                console.error("[WebRTC WS] Failed to add ICE candidate:", err);
              }
            } else {
              // Queue candidate for later
              iceCandidatesQueue.current.push(candidate);
              console.log("[WebRTC WS] Queued ICE candidate (remoteDescription not set yet):", candidate);
            }
            break;
          }
          case "hangup": {
            console.log(`[WebRTC WS] Received HANGUP from sender: ${audioPayload.sender_id}`);
            //& handle hangup state <- when peerID becomes unavailable -> hangup the call // asking reciever client to hang up the calling connection and media devices as peer has been disconnected
            handleHangup();
            break;
          }
        }
      });

      return () => {
        if (unsubscribe) {
          console.log("unsubcribing audio calling ws connection"); // cleanup ws connection from dead loose ends -> unsubcribe happens when ws connection is lost or gracefully shut downed
          unsubscribe();
        }
      };
    }, [subscribeNotifications, passedCurrentUserID]); //only re render calling service if current userID get alernated unexpectedly

    // Duration Timer management
    useEffect(() => {
      let interval = null;
      if (callState === "active") {
        setSeconds(0);
        setFinalDuration("");
        interval = setInterval(() => {
          setSeconds((prev) => prev + 1);
        }, 1000);
      } else if (callState === "hangup") {
        const mins = Math.floor(seconds / 60)
          .toString()
          .padStart(2, "0");
        const secs = (seconds % 60).toString().padStart(2, "0");
        setFinalDuration(`${mins}:${secs}`);

        // Reset back to idle state after 3 seconds
        const timeout = setTimeout(() => {
          setCallState("idle");
          setSeconds(0);
        }, 3500);

        return () => clearTimeout(timeout);
      } else if (callState === "idle") {
        setSeconds(0);
      }

      return () => {
        if (interval) clearInterval(interval);
      };
    }, [callState]);

    // fires up calling and sending offer to the reciever
    const IntialiseCalling = async (recieverID) => {
      // 1. sending a constructed payload via ws connection to let reader's publisher publish the payload to the exchange ->
      // 2. as always consumer keep chekcing for the incoming delivery in the exchange, if there is any delivery stamped for the reciever with type "offer" ->
      // 3. redirects the constructed payload with audio_payload to the hub's audio chan which ->
      // 4. redirects the incoming payload with attached peer{callingPeer~partner} to the reciever which <- then it's on reciever to either answer or decline

      // we can store anything on the set ref variable's current state <- keeps audioPayload/conn intact
      targetPeerID.current = recieverID; // so setting the target user to be someone for whom this fnct would be invoked to send 'offer'
      try {
        //! Calling flow
        // current state would store essential things
        console.log("[WebRTC Call] Initializing calling to recieverID:", recieverID);
        setCallState("calling");

        // ! pre-activate the remote audio element with the user click gesture to bypass autoplay policy blocks
        const el = document.getElementById("remote-hidden-audio-element");
        if (el) {
          el.muted = false;
          el.volume = 1.0;
          el.play().catch((err) => console.log("[WebRTC Call] Pre-play gesture activation:", err.message));
        }

        // 1. grab mic/permissions first
        // setting localStream <- senderSide permission for audio access
        console.log("[WebRTC Call] Requesting local microphone stream...");
        localStream.current = await navigator.mediaDevices.getUserMedia({
          // mediaDevices are readyOnly props <- read what client has provided to the browser
          audio: true,
          video: false,
        });

        const checktrack = localStream.current?.getTracks()[0];
        console.log({
          enabled: checktrack?.enabled,
          muted: checktrack?.muted,
          readyState: checktrack?.readyState,
          kind: checktrack?.kind
        })

        console.log("[WebRTC Call] Local microphone acquired, stream ID:", localStream.current?.id);
        localStream.current.getTracks().forEach((track, i) => {
          logTrackDetails(`Local [${i}]`, track);
        });

        // 2. initialize rtcConnection
        // setting peerConnection to store rtcConnection for audio connection
        peerConnection.current = new RTCPeerConnection(rtcConfig); // passing in stun servers to open peerConnection,which -> is store in it's current state
        setupPeerConnectionListeners(peerConnection.current);

        // 3. attach~connect the permitted microphone to reciever side {peer}
        localStream.current.getTracks().forEach((track) => {
          console.log("[WebRTC Call] Adding local track to connection:", track.id);
          peerConnection.current.addTrack(track, localStream.current);
        });

        peerConnection.current.onicecandidate = (e) => {
          // since up till now -> both peer would have been connected to the peerConn built from rtcConnection
          if (e.candidate) {
            // if they exists, in rtcConnection clients exists as candidates
            // sending 'ice-candidate' type of payload to the ws connection so that to be published and consumed by the reciever with peerID being the calling partner who had made 'offer' request
            console.log("[WebRTC Call] Generated local ICE Candidate (Caller):", {
              candidate: e.candidate.candidate,
              sdpMid: e.candidate.sdpMid,
              sdpMLineIndex: e.candidate.sdpMLineIndex
            });
            if (sendNotifications) {
              // if ws.send is available
              const payload = {
                sender_id: Number(passedCurrentUserID),
                reciever_id: Number(targetPeerID.current), // stored in its current state, we can use ref to store things dynamically throughout the application
                type: "ice-candidate",
                audio_payload_only: e.candidate,
              };

              sendNotifications(payload); // this would send this type of payload to the client's reader which would be consumed with attached peerID as senderID
              console.log(peerConnection.current?.getSenders(), "sender");
            }
          } else {
            console.log("[WebRTC Call] ICE Candidate Gathering complete (Caller)");
          }
        };
        // streaming incoming stream from peerRtcConnection
        peerConnection.current.ontrack = (e) => {
          const track = e.track; // always use the track directly, not e.streams[0] which can be muted/empty on first fire
          // on track property gives us remotely recieved stream <- in streams array being the 0th as first element being the recieved stream
          const remoteAudioStream = e.streams[0] || new MediaStream([track]); // fallback: build a stream from the track itself if streams[0] is missing

          console.log("[WebRTC Call] ontrack event received:", {
            streamsCount: e.streams.length,
            trackKind: track.kind,
            trackId: track.id,
            trackMuted: track.muted,
            trackReadyState: track.readyState
          });
          logTrackDetails("Remote (Call)", track);

          let remoteHiddenAudioElement = document.getElementById(
            "remote-hidden-audio-element",
          ); // ohhh, this would be played in hidden side but hearble track

          // if (!remoteHiddenAudioElement) {
          //   console.log("[WebRTC Call] Creating remote-hidden-audio-element");
          //   remoteHiddenAudioElement = document.createElement("audio");
          //   remoteHiddenAudioElement.id = "remote-hidden-audio-element";
          //   document.body.appendChild(remoteHiddenAudioElement);
          // }

          // if element exists and cause it would be hidden, sourcing the recieved stream from peerConnection rtc to source in from e.streams[at0thPlace]
          if (remoteHiddenAudioElement) {
            // ! setting {src} of this hidden el to play this audio track steam from the candidate
            if (remoteHiddenAudioElement.srcObject !== remoteAudioStream) {
              console.log("[WebRTC Call] Setting remoteHiddenAudioElement.srcObject to remoteAudioStream");
              remoteHiddenAudioElement.srcObject = remoteAudioStream;
            }
            remoteHiddenAudioElement.autoplay = true;
            remoteHiddenAudioElement.playsInline = true;
            remoteHiddenAudioElement.muted = false; // explicitly unmute - browser can default to muted for autoplay policy

            // ! critical: ontrack fires BEFORE ICE connects - track starts muted, play() hits silence.
            // ! bind onunmute to restart playback the moment ICE connects and audio starts flowing
            track.onunmute = () => {
              console.log("[WebRTC Call] Track UNMUTED (ICE connected, audio flowing) - restarting play()");
              remoteHiddenAudioElement.play().catch((err) => {
                console.error("[WebRTC Call] Audio playback failed on unmute:", err);
              });
            };

            console.log("[WebRTC Call] Playing remote audio...");
            remoteHiddenAudioElement.play()
              .then(() => {
                console.log("[WebRTC Call] Playback started successfully.");
              })
              .catch((err) => {
                console.error("[WebRTC Call] Audio playback failed:", err);
              });
          }
        };

        // 6. if both parties are okay, create a sdp offer,send to ws handler
        console.log("[WebRTC Call] Creating offer...");
        const offerRequest = await peerConnection.current.createOffer({ offerToReceiveAudio: true }); // starts a remote rtc connection to the peer, explicitly requesting bidirectional audio
        console.log("[WebRTC Call] Create offer SUCCESS. Setting local description...");
        await peerConnection.current.setLocalDescription(offerRequest); //* sets the session description in peerConn to be this offer request between candidates
        console.log("[WebRTC Call] Set local description SUCCESS");

        if (sendNotifications) {
          // if ws.send is available
          // @ note -> every audio_payload_only is generated from peerConn methods -> all we did is checking the flow and sending candidate req with e.candidate from the peer connection to send that tyoe of payload
          // @ - like this, also attaching payload "offer" created from same peerConnection n <=- every needy thing is created from peerConnection which holds all the things for setting up calling connection
          // @ - which the peerConnection is created from rtc
          const payload = {
            sender_id: Number(passedCurrentUserID),
            reciever_id: Number(targetPeerID.current), // stored in its current state, we can use ref to store things dynamically throughout the application
            type: "offer",
            audio_payload_only: offerRequest,
          };

          console.log("[WebRTC Call] Sending offer notification via WS to targetPeerID:", targetPeerID.current);
          sendNotifications(payload);
        } else {
          console.warn("[WebRTC Call] sendNotifications is not available to send OFFER");
        }
      } catch (err) {
        console.error("[WebRTC Call] Failed to start calling:", err);
        triggerHangup(true);
      }
    };

    const getFormattedTime = () => {
      const mins = Math.floor(seconds / 60)
        .toString()
        .padStart(2, "0");
      const secs = (seconds % 60).toString().padStart(2, "0");
      return `${mins}:${secs}`;
    };

    if (callState === "idle") {
      return (
        <audio
          id="remote-hidden-audio-element"
          autoPlay
          playsInline
          style={{ display: "none" }}
        />
      );
    }

    const getStatusText = () => {
      switch (callState) {
        case "calling":
          return "Calling...";
        case "incoming":
          return "Incoming Call";
        case "active":
          return "Connected";
        case "hangup":
          return "Call Ended";
        default:
          return "";
      }
    };

    const peerLabel = targetPeerID.current || incomingCallFrom || "User";

    return (
      <div className="audiocall-overlay">
        <div className="audiocall-card">
          <audio
            id="remote-hidden-audio-element"
            autoPlay
            playsInline
            style={{ display: "none" }}
          />
          {/* Pulsating avatar placeholder */}
          <div
            className={`audiocall-avatar ${callState === "calling" || callState === "active"
              ? "pulse-calling"
              : callState === "incoming"
                ? "pulse-incoming"
                : ""
              }`}
          >
            {String(peerLabel).substring(0, 2).toUpperCase()}
          </div>

          <h3 className="audiocall-name">User #{peerLabel}</h3>

          <p
            className={`audiocall-status ${callState === "active" ? "is-active" : "is-pending"}`}
          >
            {getStatusText()}
          </p>

          {/* Dynamic call timer */}
          {callState === "active" && (
            <div className="audiocall-timer">{getFormattedTime()}</div>
          )}

          {/* Final Talk duration display */}
          {callState === "hangup" && finalDuration && (
            <div className="audiocall-duration">
              Call Duration:{" "}
              <span className="audiocall-duration-value">{finalDuration}</span>
            </div>
          )}

          {/* Action Controls */}
          <div className="audiocall-actions">
            {callState === "incoming" ? (
              <>
                {/* Accept button */}
                <button
                  onClick={acceptCall}
                  className="audiocall-btn audiocall-btn-accept"
                >
                  📞
                </button>
                {/* Decline button */}
                <button
                  onClick={() => triggerHangup(true)}
                  className="audiocall-btn audiocall-btn-decline"
                >
                  ✖
                </button>
              </>
            ) : (
              callState !== "hangup" && (
                /* Hangup action button */
                <button
                  onClick={() => triggerHangup(true)}
                  className="audiocall-btn audiocall-btn-decline"
                >
                  🛑
                </button>
              )
            )}
          </div>
        </div>
      </div>
    );
  },
);

AudioCalling.displayName = "AudioCalling";
export default AudioCalling;
