import socket
import logging
import signal
from common.bet_processing import process_bets, obtain_winners_documents

# Fix header and bet sizes
MessageHeaderSize = 8  # 4 bytes agencyNumber + 2 bytes HeaderType + 2 bytes num_bets
BetSize = 170  # 64 + 64 + 32 + 8 + 2 (matches client)
betMessageType = 1
notificationMessageType = 2
requestWinnerMessageType = 3

STORAGE_FILEPATH = "./bets.csv"
ACK_ANSWER = 1
WINNERS_ANSWER = 2
ANSWER_HEADER_SIZE = 4 

class Server:
    def __init__(self, port, listen_backlog):
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._running = True
        self._notified_agencies = set()
        self.winners = {}
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

    def _send_ack(self, client_sock):
        try:
            response = ACK_ANSWER.to_bytes(2, byteorder='big') + (0).to_bytes(2, byteorder='big')
            client_sock.sendall(response)
        except Exception as e:
            logging.error(f"action: send_ack | result: fail | error: {e}")

    def _handle_incoming_messages(self, client_sock):
        """
        Reads data from the client socket.
        Returns: (agency_id, num_bets, bets_data) or raises an exception.
        """
        try:
            header = client_sock.recv(MessageHeaderSize)

            if not header:
                logging.info("action: receive_message | result: success")
                return None
            
            if len(header) < MessageHeaderSize:
                raise ValueError("Incomplete header received")

            return self._handle_message(header, client_sock)
        except Exception as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
            return None

    def _handle_message(self, header, client_sock):
        agency_id = int.from_bytes(header[0:4], byteorder='big')
        message_type = int.from_bytes(header[4:6], byteorder='big')
        num_bets = int.from_bytes(header[6:8], byteorder='big')
        if( 1 > message_type > 3):
            raise ValueError("Invalid message type")
        elif( message_type == betMessageType):
            return self._handle_bets_message(agency_id, num_bets, client_sock)
        elif( message_type == notificationMessageType):
            return self._handle_notification_message(agency_id, client_sock)
        elif( message_type == requestWinnerMessageType):
            return self._handle_winners_request_message(agency_id, client_sock)

    def _handle_bets_message(self, agency_id, num_bets, client_sock):
        """
        Decodes and processes the bet data.
        """
        try:
            total_size = num_bets * BetSize
            bets_data = b""

            while len(bets_data) < total_size:
                chunk = client_sock.recv(min(1024, total_size - len(bets_data)))
                if not chunk:
                    raise ValueError("Connection closed before full message was received")
                bets_data += chunk
            
            process_bets(agency_id, num_bets, bets_data)
            logging.info(f"action: apuesta_recibida | result: success | cantidad: {num_bets}")
            self._send_ack(client_sock)
            return True

        except Exception as e:
            logging.error(f"action: process_bets | result: fail | error: {e}")
            logging.error(f"action: apuesta_recibida | result: fail | cantidad: ${num_bets}")
            return False
    
    def _handle_notification_message(self, agency_id, client_sock):
        """
        Notifies the agency that the bets have been processed.
        """
        try:
            logging.info(f"action: notificacion_recibida | result: success | agencia: {agency_id}")
            self._notified_agencies.add(agency_id)
            if len(self._notified_agencies) == 5:
                self.realizar_sorteo()
            self._send_ack(client_sock)
            return True
        except Exception as e:
            logging.error(f"action: notificacion_recibida | result: fail | error: {e}")
            return False
        
    def realizar_sorteo(self):
        """
        Selects the winners and sends the notifications.
        """
        try:

            winners = obtain_winners_documents()
            for agency_id in self._notified_agencies:
                for winner in winners:
                    if winner.agency == agency_id:
                        self.winners[agency_id] = winner
            logging.info("action: sorteo_realizado | result: success")
            return True
        except Exception as e:
            logging.error(f"action: sorteo_realizado | result: fail | error: {e}")
            return False    
        
    def _send_winners(self, client_sock, winners_list):
        try:
            payload = b"".join(winner.to_bytes(2, byteorder='big') for winner in winners_list)
            response = WINNERS_ANSWER.to_bytes(2, byteorder='big') + len(payload).to_bytes(2, byteorder='big') + payload
            client_sock.sendall(response)
        except Exception as e:
            logging.error(f"action: send_winners | result: fail | error: {e}")

    def _handle_winners_request_message(self, agency_id, client_sock):
        """
        Sends the winners to the agency.
        """
        try:
            logging.info(f"action: solicitud_ganadores | result: success | agencia: {agency_id}")
            if agency_id in self.winners:
                winners_list = "\n".join(self.winners[agency_id]).encode('utf-8')
                self._send_winners(client_sock, winners_list)
            else:
                response = WINNERS_ANSWER.to_bytes(2, byteorder='big') + (0).to_bytes(2, byteorder='big')
                client_sock.sendall(response)
                
            return True
        except Exception as e:
            logging.error(f"action: solicitud_ganadores | result: fail | error: {e}")
            return False

    def __handle_client_connection(self, client_sock):
        """
        Handles the client connection, delegating reading and processing.
        """
        try:
            while self._running:  
                if self._handle_incoming_messages(client_sock) == False:
                    break

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
