export const useRequest = () => {
  const API_URL = "http://localhost:8080/";

  const fetchNft = async (endpoint: string) => {
    const response = await fetch(endpoint, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
    });
    let data = await response.json();
    data = data.data;
    if (!data) {
      return [];
    }

    for (let i = 0; i < data.length; i++) {
      const element = data[i];
      const uri = element.URI;
      console.log(uri);

      if (uri) {
        const metadata = await fetch(uri);
        const metadataJson = await metadata.json();
        element.metadata = metadataJson;
      }
    }

    if (!data) {
      return [];
    }
    return data;
  };

  const getCollectionNft = async (collection: string) => {
    const endpointUrl = API_URL + "collection/" + collection;

    return await fetchNft(endpointUrl);
  };

  const getCollectionHistory = async (collection: string) => {
    const endpointUrl = API_URL + "collection/history/" + collection;

    const response = await fetch(endpointUrl, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
    });

    let data = await response.json();
    data = data.data;
    console.log(data);
    if (!data) {
      return [];
    }
    return data;
  };

  const getAddressNft = async (address: string) => {
    const endpointUrl = API_URL + "address/" + address;
    const res = await fetchNft(endpointUrl);
    console.log(res);

    return res;
  };

  return { getAddressNft, getCollectionNft, getCollectionHistory };
};
