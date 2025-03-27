### Ejercicio N°4:
odificar la lógica de negocio tanto de los clientes como del servidor para nuestro nuevo caso de uso.

#### Cliente
Emulará a una _agencia de quiniela_ que participa del proyecto. Existen 5 agencias. Deberán recibir como variables de entorno los campos que representan la apuesta de una persona: nombre, apellido, DNI, nacimiento, numero apostado (en adelante 'número'). Ej.: `NOMBRE=Santiago Lionel`, `APELLIDO=Lorca`, `DOCUMENTO=30904465`, `NACIMIENTO=1999-03-17` y `NUMERO=7574` respectivamente.

Los campos deben enviarse al servidor para dejar registro de la apuesta. Al recibir la confirmación del servidor se debe imprimir por log: `action: apuesta_enviada | result: success | dni: ${DNI} | numero: ${NUMERO}`.

#### Servidor
Emulará a la _central de Lotería Nacional_. Deberá recibir los campos de la cada apuesta desde los clientes y almacenar la información mediante la función `store_bet(...)` para control futuro de ganadores. La función `store_bet(...)` es provista por la cátedra y no podrá ser modificada por el alumno.
Al persistir se debe imprimir por log: `action: apuesta_almacenada | result: success | dni: ${DNI} | numero: ${NUMERO}`.

#### Comunicación:
Se deberá implementar un módulo de comunicación entre el cliente y el servidor donde se maneje el envío y la recepción de los paquetes, el cual se espera que contemple:
* Definición de un protocolo para el envío de los mensajes.
* Serialización de los datos.
* Correcta separación de responsabilidades entre modelo de dominio y capa de comunicación.
* Correcto empleo de sockets, incluyendo manejo de errores y evitando los fenómenos conocidos como [_short read y short write_](https://cs61.seas.harvard.edu/site/2018/FileDescriptors/).

### Ejercicio N°5 explicación de solución:
Se creó un protocolo para la comunicación entre el cliente y el servidor.
Este consiste en el envío de mensajes encodeados a bytes a través del socket TCP. Los mensajes provenientes del cliente tienen la siguiente estructura:
## Header
Tiene un tamaño fijo de 6 bytes, contiene:
-Número de agencia (4 bytes) representado como un entero de 32 bits en formato big-endian.
-Número de apuestas (2 bytes) representado como un entero sin signo de 16 bits en formato big-endian.

## Body
Este contiene la información de las apuestas. Cada apuesta tiene un tamaño fijo de 146 bytes. 
# Campos de la apuesta:
-Nombre ( 64 bytes) se trata de un string que se rellena con espacios cuando es mas corto.
-Apellido ( 64 bytes) al igual que el nombre, un string que se rellena con espacios cuando es mas corto.
-Documento ( 8 bytes), rellenado con espacios cuando es más corto.
-Fecha de Nacimiento (8 bytes), en formato ISO, YYYYMMDD
-Número apostado (2 bytes), entero sin signo de 16 bits.

Por lo tanto el tamaño total de un mensaje es Header(6 bytes) + Tamaño de una apuesta