import {
  Box,
  Button,
  Flex,
  Input,
  Link,
  SimpleGrid,
  Table,
  TableCaption,
  TableContainer,
  Tag,
  Tbody,
  Td,
  Tfoot,
  Th,
  Thead,
  Tr,
} from "@chakra-ui/react";
import React, { useEffect, useState } from "react";
import { Fade } from "../elements/Fade";
import { NftListItem } from "../elements/NftListItem";
import { NoListItems } from "../elements/NoListItems";
import { useRequest } from "../../hooks/useRequest";
import { set } from "lodash";
import { Loading } from "../elements/Loading";

const cropInTheMiddle = (str: string, head: number, tail: number) => {
  if (str.length <= head + tail) {
    return str;
  }
  return (
    str.substr(0, head) + "..." + str.substr(str.length - tail, str.length)
  );
};

const Component: React.FC = () => {
  const [allTokens, setAllTokens] = useState<Array<any>>([]);
  const [address, setAddress] = useState<string>(
    "0x33084a2a5e90622033caac1fe1aa0ed2de41cf4b"
  );
  const [loading, setLoading] = useState<boolean>(false);
  const [history, setHistory] = useState<Array<any>>([]);
  const { getCollectionNft, getCollectionHistory } = useRequest();

  async function fetchNft() {
    setLoading(true);
    setAllTokens(
      ((await getCollectionNft(address)) as any).map((item: any) => {
        return {
          owner: item.Owner,
          metadata: {
            name: item.metadata?.name ?? "",
            image: item.metadata?.image ?? "",
          },
        };
      })
    );
    setLoading(false);
  }

  async function fetchHistory() {
    setHistory(await getCollectionHistory(address));
  }

  useEffect(() => {
    fetchNft();
    fetchHistory();
  }, []);

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
              fetchHistory();
            }}
          >
            Search
          </Button>
        </Flex>
        {loading && (
          <Flex marginY={"1.5rem"}>
            <Loading />
          </Flex>
        )}
        {!loading && (
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
        )}
        <TableContainer
          sx={{
            borderRadius: "1rem",
            padding: "1rem",
            boxShadow: "0 0 1rem rgba(0,0,0,0.05)",
          }}
        >
          <Table variant="simple">
            <TableCaption>Collection Activity</TableCaption>
            <Thead>
              <Tr>
                <Th>Time</Th>
                <Th>Tag</Th>
                <Th>TxHash</Th>
                <Th>From</Th>
                <Th>To</Th>
                <Th>Amount</Th>
              </Tr>
            </Thead>
            <Tbody>
              {history.map((item, index) => {
                const date = new Date(item.Timestamp * 1000);
                const explorerLinkTx = `https://lineascan.build/tx/${item.TxHash}`;
                return (
                  <Tr key={index}>
                    <Td>{date.toLocaleString()}</Td>
                    <Td>
                      <Tag variant="solid" colorScheme="teal">
                        {item.Tag}
                      </Tag>
                    </Td>
                    <Td>
                      <Link href={explorerLinkTx} color={"blue.400"}>
                        {cropInTheMiddle(item.TxHash, 8, 8)}
                      </Link>
                    </Td>
                    <Td>
                      <Link
                        href={`https://lineascan.build/address/${item.FromAddr}`}
                        color={"blue.400"}
                      >
                        {cropInTheMiddle(item.FromAddr, 8, 8)}
                      </Link>
                    </Td>
                    <Td>
                      <Link
                        href={`https://lineascan.build/address/${item.ToAddr}`}
                        color={"blue.400"}
                      >
                        {cropInTheMiddle(item.ToAddr, 8, 8)}
                      </Link>
                    </Td>
                    <Td>{Number(item.Value) / 10 ** 18} ETH</Td>
                  </Tr>
                );
              })}
            </Tbody>
          </Table>
        </TableContainer>
      </Box>
    </Fade>
  );
};

export { Component as Collection };
