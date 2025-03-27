import socket
import logging
import signal
from common.utils import Bet, store_bets

# Fix header and bet sizes
MessageHeaderSize = 6  # 4 bytes agencyNumber + 2 bytes num_bets
BetSize = 146  # 64 + 64 + 8 + 8 + 2 (matches client)
NAME_SIZE = 64
SURNAMES_SIZE = 64
DOCUMENT_SIZE = 8
BIRTHDATE_SIZE = 8
CHOSEN_NUMBER_SIZE = 2

STORAGE_FILEPATH = "./bets.csv"

class Server:
    def __init__(self, port, listen_backlog):
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._running = True
        signal.signal(signal.SIGTERM, self.__handle_sigterm)

    def __handle_sigterm(self, signum, frame):
        logging.info("action: handle_signal | result: success")
        self._running = False

    def shutdown(self):
        logging.info("action: shutdown_server | result: in_progress")
        if self._server_socket:
            try:
                self._server_socket.close()
                logging.info("action: close_server_socket | result: success")
            except Exception as e:
                logging.error(f"action: close_server_socket | result: fail | error: {e}")
        logging.info("action: shutdown_server | result: success")

    def _read_client_data(self, client_sock):
        """
        Reads data from the client socket.
        Returns: (agency_id, num_bets, bets_data) or raises an exception.
        """
        try:
            # Read the header (6 bytes)
            header = client_sock.recv(MessageHeaderSize)

            if not header:
                logging.info("action: receive_message | result: success")
                return None
            
            if len(header) < MessageHeaderSize:
                raise ValueError("Incomplete header received")

            agency_id = int.from_bytes(header[0:4], byteorder='big')
            num_bets = int.from_bytes(header[4:6], byteorder='big')

            logging.debug(f"Received header: agency_id={agency_id}, num_bets={num_bets}")

            total_size = num_bets * BetSize
            msg = b""

            # Read the bets
            while len(msg) < total_size:
                chunk = client_sock.recv(min(1024, total_size - len(msg)))
                if not chunk:
                    raise ValueError("Connection closed before full message was received")
                msg += chunk

            return agency_id, num_bets, msg

        except Exception as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
            return None


    def _process_client_data(self, agency_id, num_bets, bets_data):
        """
        Decodes and processes the bet data.
        """
        try:
            bets = self.decode_bets(agency_id, num_bets, bets_data)
            store_bets(bets)

            logging.info(f"action: apuesta_recibida | result: success | cantidad: {num_bets}")
            return True
        except Exception as e:
            logging.error(f"action: process_bets | result: fail | error: {e}")
            logging.error(f"action: apuesta_recibida | result: fail | cantidad: ${num_bets}")
            return False


    def __handle_client_connection(self, client_sock):
        """
        Handles the client connection, delegating reading and processing.
        """
        try:
            while self._running:  # 🔄 Loop to handle multiple messages per client connection
                result = self._read_client_data(client_sock)
                if not result:
                    break

                agency_id, num_bets, bets_data = result
                success = self._process_client_data(agency_id, num_bets, bets_data)

                # Send acknowledgment
                client_sock.sendall(b"OK" if success else b"ERROR")

        except Exception as e:
            logging.error(f"action: handle_client | result: fail | error: {e}")
        finally:
            client_sock.close()

    def __accept_new_connection(self):
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c

    def run(self):
        logging.info("action: server_start | result: in_progress")
        try:
            while self._running:
                try:
                    self._server_socket.settimeout(1.0)
                    client_sock = self.__accept_new_connection()
                    if client_sock:
                        self.__handle_client_connection(client_sock)
                except socket.timeout:
                    continue
                except Exception as e:
                    if self._running:
                        logging.error(f"action: server_loop | result: error | error: {e}")
        except Exception as e:
            logging.error(f"action: server_loop | result: error | error: {e}")
        finally:
            self.shutdown()
            logging.info("action: server_shutdown | result: success")

    def decode_bets(self, agency_id, num_bets, message: bytes) -> list[Bet]:
        """
        Decode a binary message into bets.
        """
        bets = []
        offset = 0

        for _ in range(num_bets):
            # Read FirstName (64 bytes)
            first_name = message[offset:offset+NAME_SIZE].split(b'\0', 1)[0].decode('utf-8').strip()
            offset += NAME_SIZE

            # Read LastName (64 bytes)
            last_name = message[offset:offset+SURNAMES_SIZE].split(b'\0', 1)[0].decode('utf-8').strip()
            offset += SURNAMES_SIZE

            # Read Document (8 bytes)
            document = message[offset:offset+DOCUMENT_SIZE].split(b'\0', 1)[0].decode('utf-8').strip()
            offset += DOCUMENT_SIZE

            # Read birthdate (8 bytes) and convert to YYYY-MM-DD format
            date_str = message[offset:offset+BIRTHDATE_SIZE].decode('utf-8')
            birthdate = f"{date_str[:4]}-{date_str[4:6]}-{date_str[6:8]}"
            offset += BIRTHDATE_SIZE

            # Read number (2 bytes) instead of 8
            number = int.from_bytes(message[offset:offset+CHOSEN_NUMBER_SIZE], byteorder='big')
            offset += CHOSEN_NUMBER_SIZE  # Updated to match client

            # Create Bet object
            bet = Bet(str(agency_id), first_name, last_name, document, birthdate, str(number))
            bets.append(bet)

        return bets
