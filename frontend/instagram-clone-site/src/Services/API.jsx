import { useEffect, useState } from "react";

// method to fetch profile data - from user struct
export function FetchProfileData(url, reqMethod) {
  const [data, setData] = useState({});
  const token = localStorage.getItem("token");
  useEffect(() => {
    async () => {
      try {
        const req = await fetch(url, {
          method: reqMethod,
          headers: {
            Authorization: token, // direct define Auth header in braces
          },
        });

        const response = await req.json();

        //  if fetch\response struct's bool req was not successfull
        if (req.ok || !response.OK) {
          console.log("error fetching data :", response.Status);
        }

        return response;
      } catch (err) {
        console.log(err);
      }
    };
  }, []);
}
