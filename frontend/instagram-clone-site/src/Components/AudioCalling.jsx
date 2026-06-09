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

    // stun servers
    const rtcConfig = {
      // type of AudioPayload sending for calling config to determine things when request is either<- "offer" or "answer"

      // must follow this native configuration naming conventions
      iceServers: [{ urls: `stun:stun.l.google.com:19302` }],
    };

    const outletContext = useOutletContext();
    const sendNotifications =
      propSendNotifications || outletContext?.sendNotifications;
    const subscribeNotifications =
      propSubscribeNotifications || outletContext?.subscribeNotifications;

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
        if (!incomingOfferSdp) return;

        // peer connection is stored in current state <- created by new RTCPeerConnection(passingInIceStunServersConfig)
        peerConnection.current = new RTCPeerConnection(rtcConfig); // just remember everything is stored in current state by the use of ref

        // before setting up connection,adding this to candidate for ipLookup by stun servers and candidate path matching for interconnection
        peerConnection.current.onicecandidate = (e) => {
          //**  sending payload of audio_type payload of context - e.candidate => if e.candidate exists in peerConn. here --
          if (e.candidate && sendNotifications) {
            sendNotifications({
              sender_id: Number(passedCurrentUserID),
              reciever_id: Number(incomingCallFrom),
              type: "ice-candidate",
              audio_payload_only: e.candidate,
            });
          }
        };

        // streaming incoming stream from peerRtcConnection
        peerConnection.current.ontrack = (e) => {
          const remoteAudioStream = e.streams[0]; // on track property gives us remotely recieved stream <- in streams array being the 0th as first element being the recieved stream
          let remoteHiddenAudioElement = document.getElementById(
            "remote-hidden-audio-element",
          );

          if (!remoteHiddenAudioElement) {
            remoteHiddenAudioElement = document.createElement("audio");
            remoteHiddenAudioElement.id = "remote-hidden-audio-element";
            document.body.appendChild(remoteHiddenAudioElement);
          }

          // if element exists and cause it would be hidden, sourcing the recieved stream from peerConnection rtc to source in from e.streams[at0thPlace]
          if (remoteHiddenAudioElement) {
            // ! setting {src} of this hidden el to play this audio track steam from the candidate
            remoteHiddenAudioElement.srcObject = remoteAudioStream;
            remoteHiddenAudioElement.autoplay = true;
            remoteHiddenAudioElement.playsInline = true;
            remoteHiddenAudioElement.play().catch((err) => {
              console.error("Audio playback failed:", err);
            });
          }
        };

        // setting session description to be from recieved 'offer' payload from ws connection
        await peerConnection.current.setRemoteDescription(
          new RTCSessionDescription(incomingOfferSdp),
        );

        // Process queued candidates
        if (iceCandidatesQueue.current.length > 0) {
          console.log(
            `Processing ${iceCandidatesQueue.current.length} queued ICE candidates on accept`,
          );
          for (const candidate of iceCandidatesQueue.current) {
            try {
              await peerConnection.current.addIceCandidate(
                new RTCIceCandidate(candidate),
              );
            } catch (err) {
              console.error("Failed to add queued ICE candidate:", err);
            }
          }
          iceCandidatesQueue.current = [];
        }

        // get his mics and all and store in localStream as streamingMic
        localStream.current = await window.navigator.mediaDevices.getUserMedia({
          audio: true,
          video: false,
        });

        // add local tracks
        localStream.current.getTracks().forEach((track) => {
          peerConnection.current.addTrack(track, localStream.current);
        });

        // now sending answer payload to the caller with sdp block and audio_payload, this time sending 'answer' payload,
        const createdAnsSdpPayload =
          await peerConnection.current.createAnswer(); //sdp answer audio payload, sending to caller with type "answer" for connection
        await peerConnection.current.setLocalDescription(createdAnsSdpPayload);

        // * every offer/ans/candidate_req payload is sent to ws, attaching audio_payload based off context of outbound requests
        if (sendNotifications) {
          sendNotifications({
            sender_id: Number(passedCurrentUserID),
            reciever_id: Number(targetPeerID.current), // senderId - id of the user whose 'offer' request was intercepted and in response 'answer' is created for him wiht ans audio_payload attached inside
            type: "answer", // sending ans payload to the ws with type being "answer" -> by unmarshaling it would know what is incoming audio_payload and what needs to be published & shipped to consume by the reciever
            audio_payload_only: createdAnsSdpPayload,
          });
        }

        setCallState("active");
      } catch (error) {
        console.error("Failed to accept incoming call:", error);
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
            targetPeerID.current = audioPayload.sender_id; // sender_id is what supplied by hub from reciever attaching -> the senderID ;reciever sends the sender id as peerID
            setIncomingCallFrom(audioPayload.sender_id);
            setIncomingOfferSdp(audioPayload.audio_payload_only);
            setCallState("incoming");
            break;
          }
          case "answer": {
            // &when call is either approved or not <- 'answering' call => reciever gets 'answer''s audio_payload
            if (peerConnection.current) {
              await peerConnection.current.setRemoteDescription(
                new RTCSessionDescription(audioPayload.audio_payload_only), // opening rtcConnection from recieved audioPayload to the reciever
              );
              setCallState("active");

              // Process queued candidates
              if (iceCandidatesQueue.current.length > 0) {
                console.log(
                  `Processing ${iceCandidatesQueue.current.length} queued ICE candidates on answer`,
                );
                for (const candidate of iceCandidatesQueue.current) {
                  try {
                    await peerConnection.current.addIceCandidate(
                      new RTCIceCandidate(candidate),
                    );
                  } catch (err) {
                    console.error("Failed to add queued ICE candidate:", err);
                  }
                }
                iceCandidatesQueue.current = [];
              }
            }
            break;
          }
          case "ice-candidate": {
            //& when both gets connected
            const candidate = audioPayload.audio_payload_only;
            if (
              peerConnection.current &&
              peerConnection.current.remoteDescription // since we store connection when answered <- if that exists
            ) {
              try {
                await peerConnection.current.addIceCandidate(
                  new RTCIceCandidate(candidate),
                );
              } catch (err) {
                console.error("Failed to add ICE candidate:", err);
              }
            } else {
              // Queue candidate for later
              iceCandidatesQueue.current.push(candidate);
              console.log("Queued ICE candidate:", candidate);
            }
            break;
          }
          case "hangup": {
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
        setCallState("calling");

        // 1. grab mic/permissions first
        // setting localStream <- senderSide permission for audio access
        localStream.current = await navigator.mediaDevices.getUserMedia({
          // mediaDevices are readyOnly props <- read what client has provided to the browser
          audio: true,
          video: false,
        });

        // 2. initialize rtcConnection
        // setting peerConnection to store rtcConnection for audio connection
        peerConnection.current = new RTCPeerConnection(rtcConfig); // passing in stun servers to open peerConnection,which -> is store in it's current state

        // 3. attach~connect the permitted microphone to reciever side {peer}
        // at this ppinte.forEach((track) => {
        //  at this point lstream would have stored the permission now we are ready to connect to the peer mics
        // })   peerConnection.current.addTrack(track,localStream.current)controlling structure -> if finds ip -> send attached ips in contructed payload via ws
        localStream.current.getTracks().forEach((track) => {
          peerConnection.current.addTrack(track, localStream.current);
        });

        peerConnection.current.onicecandidate = (e) => {
          // since up till now -> both peer would have been connected to the peerConn built from rtcConnection
          if (e.candidate) {
            // if they exists, in rtcConnection clients exists as candidates
            // sending 'ice-candidate' type of payload to the ws connection so that to be published and consumed by the reciever with peerID being the calling partner who had made 'offer' request
            if (sendNotifications) {
              // if ws.send is available
              const payload = {
                sender_id: Number(passedCurrentUserID),
                reciever_id: Number(targetPeerID.current), // stored in its current state, we can use ref to store things dynamically throughout the application
                type: "ice-candidate",
                audio_payload_only: e.candidate,
              };

              sendNotifications(payload); // this would send this type of payload to the client's reader which would be consumed with attached peerID as senderID
            }
          }
        };

        // streaming incoming stream from peerRtcConnection
        peerConnection.current.ontrack = (e) => {
          const remoteAudioStream = e.streams[0]; // on track property gives us remotely recieved stream <- in streams array being the 0th as first element being the recieved stream
          let remoteHiddenAudioElement = document.getElementById(
            "remote-hidden-audio-element",
          ); // ohhh, this would be played in hidden side but hearble track

          if (!remoteHiddenAudioElement) {
            remoteHiddenAudioElement = document.createElement("audio");
            remoteHiddenAudioElement.id = "remote-hidden-audio-element";
            document.body.appendChild(remoteHiddenAudioElement);
          }

          // if element exists and cause it would be hidden, sourcing the recieved stream from peerConnection rtc to source in from e.streams[at0thPlace]
          if (remoteHiddenAudioElement) {
            // ! setting {src} of this hidden el to play this audio track steam from the candidate
            remoteHiddenAudioElement.srcObject = remoteAudioStream;
            remoteHiddenAudioElement.autoplay = true;
            remoteHiddenAudioElement.playsInline = true;
            remoteHiddenAudioElement.play().catch((err) => {
              console.error("Audio playback failed:", err);
            });
          }
        };

        // 6. if both parties are okay, create a sdp offer,send to ws handler
        const offerRequest = await peerConnection.current.createOffer(); // starts a remote rtc connection to the peer
        await peerConnection.current.setLocalDescription(offerRequest); //* sets the session description in peerConn to be this offer request between candidates

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

          sendNotifications(payload);
        }
      } catch (err) {
        console.error("Failed to start calling:", err);
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

    if (callState === "idle") return null;

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
          {/* Pulsating avatar placeholder */}
          <div
            className={`audiocall-avatar ${
              callState === "calling" || callState === "active"
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
