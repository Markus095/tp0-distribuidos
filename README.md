### Ejercicio N°4:
Modificar servidor y cliente para que ambos sistemas terminen de forma _graceful_ al recibir la signal SIGTERM. Terminar la aplicación de forma _graceful_ implica que todos los _file descriptors_ (entre los que se encuentran archivos, sockets, threads y procesos) deben cerrarse correctamente antes que el thread de la aplicación principal muera. Loguear mensajes en el cierre de cada recurso (hint: Verificar que hace el flag `-t` utilizado en el comando `docker compose down`).

### Ejercicio N°4 explicación de la solución:
Se le agregó al servidor un nuevo parámetro: running, el cual es revisado en lugar del while true original. Se inicializa con un valor positivo, pero cambia cuando recibe el sigterm. Cuando eso sucede, se corta el loop y trata de cerrar todos los sockets y luego finaliza. 
Por otro lado, se le agregó un waitgroup al cliente para revisar que las conexiones se cierran antes de terminar el programa y un campo "done" que se usa para avisarle a todas las rutinas cuando se solicita el cierre del cliente. También se le añadió un signal handler que funciona en una rutina separada para manejar la secuencia de cerrado y la secuencia de cierre, que cierra todas las conexiones.

### Dificultades encontradas
Al principio me costó encontrar el formato apropiado para el logging, así que tuve varios commits innecesarios para probar como funcionaban distintos formatos con los tests.
Durante bastante tiempo le pasaba una cantidad erronea de parametros al handler del SIGTERM del servidor, así que el cierre no funcionaba correctamente.