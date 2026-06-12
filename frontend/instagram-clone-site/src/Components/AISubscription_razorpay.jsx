import denvergam_logo from "../assets/icons/denvergram_logo_transparent.png";
export default function AISubscription_razorpay() {
  // mocking payment integration - with test keys only

  // # registering handlers
  const rzrpayID =
    import.meta.env.VITE_RZRPAY_ACCESS_KEY_ID || "rzp_test_T0iz587aGLSrzC";
  //   const {};
  //** when fired -> it triggers the payment (in sandbox)
  const handleTestPayment = () => {
    // 1. configuring the poped up payload firstly
    const options = {
      // secret key goes here -> works as approved rzrpay client -> handles incoming payment request on their url ( opens in their sandbox)
      key: rzrpayID,

      // *PHASE 1 - onboarding information
      amount: 29900, // in paise , 100 paise = 1rupee
      currency: "INR",
      name: "ayush kumar",
      description: "1-Month DenverAI premium subscription",
      image: denvergam_logo, // site logo here

      // * PHASE 2 - UPI PAYMENT EVENT HANDLER (in their own sandbox) **//
      handler: function (res) {
        alert("🎉 transaction successfull, upgraded to premium plan 🎉");
        console.log("Payment ID Recieved:", res.razorpay_payment_id);

        // ! inside handler where we upgrade subscription  status for the client in our db
        // ** grants him the upgraded privileges ( need backend )
      },
      prefill: {
        name: "ayush kumar",
        email: "[EMAIL_ADDRESS]",
        contact: "8226928546",
      },
      notes: {
        address: "Razorpay Corporate Office",
      },
      theme: {
        color: "#1480a1ff",
      },
    };

    // ! important - All the information is feeded to rzpClient -> which handles all the transaction on their servers
    // !and opens up payment window overlay dynamically by own
    const rzpClient = window.Razorpay(options);
    rzpClient.open(); // -> it triggers the payment window (overlay)
  };
  return (
    <>
      <div style={{ padding: "20px", textAlign: "center" }}>
        <h3>Unlock Unlimited AI Capabilities</h3>
        <p>
          Test premium token assignment with a simulated ₹199 sandbox layer.
        </p>
        <button
          onClick={handleTestPayment}
          style={{
            backgroundColor: "#6366f1",
            color: "white",
            padding: "12px 24px",
            border: "none",
            borderRadius: "8px",
            cursor: "pointer",
            fontSize: "16px",
            fontWeight: "bold",
          }}
        >
          Upgrade with Test Sandbox
        </button>
      </div>
    </>
  );
}
