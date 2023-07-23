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
    return await fetchNft(endpointUrl);
  };

  return { getAddressNft, getCollectionNft, getCollectionHistory };
};
