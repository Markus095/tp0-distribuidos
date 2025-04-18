import socket
import logging
import signal
import os
from multiprocessing import Process, Manager, Lock
from common.bet_processing import process_bets, obtain_winners_documents

MESSAGE_HEADER_SIZE = 8
BET_SIZE = 146
BET_MESSAGE_TYPE = 1
NOTIFICATION_MESSAGE_TYPE = 2
REQUEST_WINNERS_MESSAGE_TYPE = 3

NUMBER_OF_CLIENTS: int = int(os.getenv("NUMBER_OF_CLIENTS", 1))

STORAGE_FILEPATH = "./bets.csv"
ACK_ANSWER = 1
NO_WINNERS_ANSWER = 2
WINNERS_ANSWER = 3
ANSWER_HEADER_SIZE = 4 

class Server:
    def __init__(self, port, listen_backlog):
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._running = True

        manager = Manager()
        self._notified_agencies = manager.list()
        self.winners = manager.dict()
        self._lock = Lock()  

        signal.signal(signal.SIGTERM, self.__handle_sigterm)
        signal.signal(signal.SIGINT, self.__handle_sigterm)
        self._processes = [] 

    def __handle_sigterm(self, signum, frame):
        logging.info("action: handle_signal | result: success")
        self._running = False
        self.shutdown()

    def shutdown(self):
        logging.info("action: shutdown_server | result: in_progress")
        if self._server_socket:
            try:
                self._server_socket.close()
                logging.info("action: close_server_socket | result: success")
            except Exception as e:
                logging.error(f"action: close_server_socket | result: fail | error: {e}")
        for process in self._processes:
            if process.is_alive():
                process.terminate()
                process.join()
        logging.info("action: terminate_processes | result: success")
        logging.info("action: shutdown_server | result: success")

    def _send_ack(self, client_sock):
        try:
            response = ACK_ANSWER.to_bytes(2, byteorder='big') + (0).to_bytes(2, byteorder='big')
            client_sock.sendall(response)
        except Exception as e:
            logging.error(f"action: send_ack | result: fail | error: {e}")

    def _send_winners(self, client_sock, winners_list):
        try:
            payload = b"".join(winner.to_bytes(4, byteorder='big') for winner in winners_list)
            response = WINNERS_ANSWER.to_bytes(2, byteorder='big') + len(payload).to_bytes(2, byteorder='big') + payload
            client_sock.sendall(response)
        except Exception as e:
            logging.error(f"action: send_winners | result: fail | error: {e}")
    
    def _send_no_winners(self, client_sock):
        try:
            response = NO_WINNERS_ANSWER.to_bytes(2, byteorder='big') + (0).to_bytes(2, byteorder='big')
            client_sock.sendall(response)
        except Exception as e:
            logging.error(f"action: send_no_winners | result: fail | error: {e}")

    def _handle_incoming_messages(self, client_sock):
        """
        Reads data from the client socket.
        """
        try:
            header = client_sock.recv(MESSAGE_HEADER_SIZE)

            if not header:  
                logging.info("action: receive_message | result: fail | reason: no_data")
                return None

            if len(header) < MESSAGE_HEADER_SIZE:
                logging.warning("action: receive_message | result: fail| reason: incomplete_header")
                return None

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
        elif( message_type == BET_MESSAGE_TYPE):
            return self._handle_bets_message(agency_id, num_bets, client_sock)
        elif( message_type == NOTIFICATION_MESSAGE_TYPE):
            return self._handle_notification_message(agency_id, client_sock)
        elif( message_type == REQUEST_WINNERS_MESSAGE_TYPE):
            return self._handle_winners_request_message(agency_id, client_sock)

    def _handle_bets_message(self, agency_id, num_bets, client_sock):
        """
        Decodes and processes the bet data.
        """
        try:
            total_size = num_bets * BET_SIZE
            bets_data = b""

            while len(bets_data) < total_size:
                chunk = client_sock.recv(min(1024, total_size - len(bets_data)))
                if not chunk:
                    raise ValueError("Connection closed before full message was received")
                bets_data += chunk
            
            process_bets(agency_id, num_bets, bets_data, self._lock)
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
            if agency_id not in self._notified_agencies:
                self._notified_agencies.append(agency_id)  # Add to shared list
                logging.info(f"action: notificacion_recibida | result: success | agencia: {agency_id} agencias_notificadas: {len(self._notified_agencies)}")
            
            if len(self._notified_agencies) == NUMBER_OF_CLIENTS:
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
            if not winners:
                logging.warning("action: obtener_ganadores | result: no_winners_found")
                return False

            for agency_id in self._notified_agencies:
                self.winners[agency_id] = [
                    int(document) for agency, document in winners if agency == agency_id and document.isdigit()
                ]
            logging.info("action: sorteo_realizado | result: success")
            return True
        except Exception as e:
            logging.error(f"action: sorteo_realizado | result: fail | error: {e}")
            return False

    def _handle_winners_request_message(self, agency_id, client_sock):
        """
        Sends the winners to the agency.
        """
        try:
            if agency_id in self.winners:
                winners_list = self.winners[agency_id]
                self._send_winners(client_sock, winners_list)
                logging.info(f"action: solicitud_ganadores | result: success | agencia: {agency_id} | cantidad: {len(winners_list)}")
            else:
                self._send_no_winners(client_sock)
                logging.info(f"action: solicitud_ganadores | result: success | agencia: {agency_id} no hay ganadores aun")
        except Exception as e:
            logging.error(f"action: solicitud_ganadores | result: fail | error: {e}")
            return False
        finally:
            client_sock.close()
        

    def __handle_client_connection(self, client_sock):
        """
        Handles the client connection, delegating reading and processing.
        This method will run in a separate process.
        """
        try:
            while self._running:
                try:
                    result = self._handle_incoming_messages(client_sock)
                    if result is None:
                        logging.info("action: handle_client | result: success | info: closing_connection")
                        break
                except ConnectionResetError:
                    logging.info("action: handle_client | result: success | reason: connection_reset")
                    break
        except Exception as e:
            logging.error(f"action: handle_client | result: fail | error: {e}")
        finally:
            client_sock.close()

    def __accept_new_connection(self):
        """
        Accepts a new client connection.
        """
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c

    def run(self):
        """
        Main server loop to accept and handle client connections in parallel.
        """
        logging.info("action: server_start | result: in_progress")
        try:
            while self._running:
                try:
                    self._server_socket.settimeout(10.0)
                    client_sock = self.__accept_new_connection()
                    if client_sock:
                        process = Process(target=self.__handle_client_connection, args=(client_sock,))
                        process.start()
                        self._processes.append(process)
                        logging.info(f"action: spawn_process | result: success | pid: {process.pid}")
                except socket.timeout:
                    logging.debug("action: server_loop | result: timeout | info: no_connections")
                except Exception as e:
                    if self._running:
                        logging.error(f"action: server_loop | result: error | error: {e}")
                for process in self._processes:
                    if not process.is_alive():
                        process.join()  # Reap the process
                        self._processes.remove(process)
        except Exception as e:
            logging.error(f"action: server_loop | result: error | error: {e}")
        finally:
            self.shutdown()
            logging.info("action: server_shutdown | result: success")
