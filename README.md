### Ejercicio N°1 explicación de solución:
Se creó un script muy sencillo de Bash, que solo se encarga de imprimir los parámetros por consola y luego enviarselos a un archivo de python. Este archivo llamado "mi-generador.py" es el encargado de:
1-Validar que son parámetros válidos para generar un nuevo archivo de compose.
2-Abrir un archivo nuevo, con el nombre especificado.
3-Agregar al servidor al archivo, siguiendo el formato del servidor de ejemplo.
4-Agregar tantos clientes como se le especificaron, siguiendo el formato anterior y asegurándose de cambiar sus ids y nombres.
5-Agregar la red a la cual están todos conectados.
6-Si pudo agregar todo y guardar el archivo con éxito, notifica al usuario y el programa termina.

## Dificultades Encontradas
-Hubo ciertas complicaciones a la hora de parsear el archivo, a veces las líneas quedaban desfasadas.


### Ejercicio N°2 explicación de solución:
Se modificó el archivo de compose (y por lo tanto el generador) para que el servidor y el/los cliente/s tengan un volumen por el cual se pueda acceder a sus respectivas configuraciones.
Además ya no se especifica su logging level desde el compose.

## Dificultades Encontradas
-Hubo cierta dificultad en encontrar el path correcto de los archivos de configuración a los volúmenes. Al principio intentaba con paths absolutos y no terminaba de entender por qué fallaba. Al usar paths relativos se solucionó.


### Ejercicio N°3 explicación de la solución:
El script funciona de la siguiente manera:
1- Crea un container temporal con netcat instalado.
2- Duerme un par de segundos para esperar la instalación.
3- Envía un mensaje al servidor a través de la red y luego captura su respuesta en una variable usando docker exec.
4- Elimina el container.
5- Compara la respuesta con el texto enviado. Si son iguales, imprime por consola que la prueba fue un éxito. Caso contrario, imprime que fue un fracaso.

## Dificultades encontradas
-Al principio intenté hacer la prueba a través de los clientes, pero terminé decidiendo que crear un container nuevo era más sencillo y práctico.
-Fue difícil comprender la lógica de conectarme a la red y obtener la ip del servidor. Tuve que leer documentación de docker network.
-Tardé un poco en entender que tenía que conectarme a través del puerto especificado en la configuración del servidor.
-Tuve que aprender a usar variables y ifs en bash, que nunca me había hecho falta hasta ahora.


### Ejercicio N°4 explicación de la solución:
Se le agregó al servidor un nuevo parámetro: running, el cual es revisado en lugar del while true original. Se inicializa con un valor positivo, pero cambia cuando recibe el sigterm. Cuando eso sucede, se corta el loop y trata de cerrar todos los sockets y luego finaliza. 
Por otro lado, se le agregó un waitgroup al cliente para revisar que las conexiones se cierran antes de terminar el programa y un campo "done" que se usa para avisarle a todas las rutinas cuando se solicita el cierre del cliente. También se le añadió un signal handler que funciona en una rutina separada para manejar la secuencia de cerrado y la secuencia de cierre, que cierra todas las conexiones.

## Dificultades encontradas
Al principio me costó encontrar el formato apropiado para el logging, así que tuve varios commits innecesarios para probar como funcionaban distintos formatos con los tests.
Durante bastante tiempo le pasaba una cantidad erronea de parametros al handler del SIGTERM del servidor, así que el cierre no funcionaba correctamente.


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

El Cliente parsea este mensaje con las variables de la apuesta que se le pasan por parámetro.
Cuando el Servidor lo recibe, almacena el ID del cliente y luego decodifica cada apuesta, usando el ID del header para el campo de la agencia. Finalmente almacena las apuestas y envía un mensaje de 2 bytes que dice ("OK) como ACK, al recibirlo el cliente termina.


### Ejercicio N°6 explicación de Solución:
El protocolo ya estaba planteado con la idea de manejar varias apuestas en un solo mensaje, así que no fue necesario modificarlo.
Se modificó al cliente para poder levantar los archivos del csv en el docker volume y enviar las apuestas en varios mensajes al servidor. Cada mensaje puede tener hasta "batch: maxAmount" apuestas, pero (generalmente el último) pueden tener menos y eso no genera ningún problema. También se lo separó en varios archivos para facilitar la lectura y poder mantener la separación de responsabilidades.
El servidor se modificó para esperar varios mensajes de un mismo cliente y para poder procesar y guardar las varias apuestas

## Dificultades
-Inicialmente moví los directorios a una carpeta "/datasets" y funcionaba bien, pero como las pruebas no lo encontraban tuve que modificar toda la lógica de la creacion de los volumenes y en lugar de pasarle un archivo ahora los clientes reciben el directorio.


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


## Dificultades
-Inicialmente no había un protocolo para diferenciar el tipo de mensaje, se hacía en base al tamaño del payload. Esto resultó muy complicado de verificar con éxito e innecesariamente complicado, así que terminé decidiendo agregar un campo extra. 

### Ejercicio N°8 Explicación de Solución:
Me decidí por una solución multiproceso debido a las limitaciones que tiene Python con multithreading.
Se modificó al servidor de forma que cada vez que recibe un cliente por el socket donde escucha conexiones nuevas, crea un proceso para manejarlo. Estos procesos se manejan de forma independiente salvo por 3 recursos: 

-El archivo donde guardan las apuestas, el cual es solo accesible al obtener el lock. Evita que condiciones de carrera y la pérdida de datos al guardar las apuestas.
-El set de agencias que ya notificaron el envío de todas sus apuestas. Es compartido y administrado por un Manager de la biblioteca multiprocessing, que asegura las sincronización de los procesos. Evita que los procesos solo sepan de la notificación de su cliente y que en consecuencia nunca lleguen al sorteo.
-El diccionario de ganadores, también administrado por el Manager. Evita que todos los procesos realicen el sorteo por separado innecesariamente, ya que es una operación costosa si se manejan grandes cantidades de apuestas.

Cuando un cliente se desconecta, el proceso que lo manejaba también finaliza y realiza un `join` con el proceso principal. Si el proceso principal recibe una señal `SIGTERM` o finaliza por algún otro motivo, primero envía una señal a los procesos secundarios para que terminen y luego realiza un `join` con ellos, evitando así la creación de procesos huérfanos.

