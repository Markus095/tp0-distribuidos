### Ejercicio N°7:

Modificar los clientes para que notifiquen al servidor al finalizar con el envío de todas las apuestas y así proceder con el sorteo.
Inmediatamente después de la notificacion, los clientes consultarán la lista de ganadores del sorteo correspondientes a su agencia.
Una vez el cliente obtenga los resultados, deberá imprimir por log: `action: consulta_ganadores | result: success | cant_ganadores: ${CANT}`.

El servidor deberá esperar la notificación de las 5 agencias para considerar que se realizó el sorteo e imprimir por log: `action: sorteo | result: success`.
Luego de este evento, podrá verificar cada apuesta con las funciones `load_bets(...)` y `has_won(...)` y retornar los DNI de los ganadores de la agencia en cuestión. Antes del sorteo no se podrán responder consultas por la lista de ganadores con información parcial.

Las funciones `load_bets(...)` y `has_won(...)` son provistas por la cátedra y no podrán ser modificadas por el alumno.

No es correcto realizar un broadcast de todos los ganadores hacia todas las agencias, se espera que se informen los DNIs ganadores que correspondan a cada una de ellas.

### Ejercicio N°7 Explicación de Solución:
El protocolo recibió varias modificaciones:
Se agregó un campo "Message Type" al Header de 2 bytes, que facilita distinguir de que mensaje se trata para el servidor.
Ahora el cliente tiene tres tipos de mensaje:
-BetsMessage, usado para enviarle apuestas al servidor
-NotificationMessage, usado para avisarle al servidor que ya terminó de enviar apuestas
-WinnersRequestMessage, usado para preguntarle al servidor si alguna de las apuestas que envió ganó.
Los dos Mensajes nuevos no tienen payload y su cantidad de apuestas siempre es cero, pero igualmente se completa este único para que el servidor siempre maneje Headers de 8 bytes.

Además, como el servidor ya no contesta solo con ACKs le cree un protocolo a sus respuestas:
## Header
El tamaño de este es de solo 4 bytes, contiene:
-TipoDeRespuesta (2 bytes), entero de 16 bits que indica si es un mensaje de ACK, un mensaje que avisa que aún no se determinaron los ganadores o un mensaje que contesta cuales son los ganadores de ese cliente.
-AmountOfWinners (2 bytes), entero de 16 bits que indica cuantos ganadores puede esperar que le lleguen al cliente.

## Body
Contiene un único campo, y solo existe si es un mensaje con los ganadores de alguna agencia. Tiene un único campo:
-Winners (8 bytes c/u), contiene los DNIs de los ganadores de esa agencia.

Por lo tanto el tamaño de una respuesta del servidor es 4 para los primeros dos casos (solo tienen Header)
y 4 + 8 * AmountOfWinners en el último.

Tanto el cliente como el servidor tuvieron cambios drásticos.
Ahora el cliente después de haber enviado todas sus apuestas envía la notificación al servidor e inmediatamente pregunta si ya están sus ganadores. En caso de respuesta negativa cierra la conexión, realiza un sleep y vuelve a conectarse para preguntar. El sleep aumenta exponencialmente con cada respuesta negativa. Si tras 5 intentos el servidor no le contestó termina la ejecución. Tras enviar los mensajes con apuestas y la notificación de haber terminado espera un ACK del servidor para continuar. Cuando envía un mensaje de solicitar ganadores espera que el servidor le conteste con los ganadores o con un aviso de que todavía no se hizo el sorteo. Al recibir una respuesta afirmativa a la cantidad de ganadores, loguea cuantas de sus apuestas ganaron y termina.

Por otro lado, el servidor ahora tiene una lógica para encodear y enviar mensajes al cliente con el que se comunica. Maneja a varios clientes de forma secuencial. Sigue manejando las apuestas enviadas como antes, pero ahora va almacenando cuantas agencias le notificaron que terminaron. Una vez que la cantidad de agencias es igual al número de clientes indicado en su variable de entorno sacada del compose, el servidor realiza el sorteo.
Para hacer el sorteo carga las apuestas que fueron guardadas a disco y posteriormente llama al has won en cada una, almacenando el DNI de los ganadores y a qué agencia pertenecen en un diccionario. Antes de realizar el sorteo contesta cualquier consulta de ganadores con un mensaje de NO_WINNERS, y tras hacerlo con los ganadores correspondientes al ID de agencia puesto en la solicitud.


### Dificultades
-Inicialmente no había un protocolo para diferenciar el tipo de mensaje, se hacía en base al tamaño del payload. Esto resultó muy complicado de verificar con éxito e innecesariamente complicado, así que terminé decidiendo agregar un campo extra. 
