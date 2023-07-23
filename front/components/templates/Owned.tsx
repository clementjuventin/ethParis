import { Box, SimpleGrid, Flex, Input, Button } from "@chakra-ui/react";
import React, { useContext, useEffect, useState } from "react";
import { NftContractContext } from "../../contexts/NftContractProvider";
import { Fade } from "../elements/Fade";
import { NftListItem } from "../elements/NftListItem";
import { NoListItems } from "../elements/NoListItems";
import { useRequest } from "../../hooks/useRequest";

const Component: React.FC = () => {
  const [allTokens, setAllTokens] = useState<Array<any>>([]);
  const [address, setAddress] = useState<string>(
    "0x8fdd8db198b292d233fb5dc191e31bebc41e1144"
  );
  const { getAddressNft } = useRequest();

  async function fetchNft() {
    setAllTokens(
      ((await getAddressNft(address)) as any).map((item: any) => {
        return {
          owner: item.Owner,
          metadata: {
            name: item.metadata?.name ?? "",
            image: item.metadata?.image ?? "",
          },
        };
      })
    );
  }

  return (
    <Fade>
      {/** Search input but */}

      <Box maxW="8xl" mx="auto">
        <Flex
          justifyContent="center"
          marginTop="1rem"
          alignItems="center"
          mx="auto"
          gap={4}
        >
          <Input
            placeholder="Search"
            onChange={(e) => {
              setAddress(e.target.value);
            }}
            value={address}
          />

          <Button
            onClick={() => {
              fetchNft();
            }}
          >
            Search
          </Button>
        </Flex>
        <SimpleGrid
          columns={{
            base: 2,
            md: 3,
            lg: 4,
            xl: 5,
            "2xl": 6,
          }}
          spacing={{ base: 3, xl: 6 }}
          py={6}
        >
          {allTokens.map((token, index) => {
            return (
              <React.Fragment key={index}>
                <NftListItem token={token} />
              </React.Fragment>
            );
          })}
          {allTokens.length == 0 && <NoListItems />}
        </SimpleGrid>
      </Box>
    </Fade>
  );
};

export { Component as Owned };
