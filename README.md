### Ejercicio N°3:
Crear un script de bash `validar-echo-server.sh` que permita verificar el correcto funcionamiento del servidor utilizando el comando `netcat` para interactuar con el mismo. Dado que el servidor es un echo server, se debe enviar un mensaje al servidor y esperar recibir el mismo mensaje enviado.

En caso de que la validación sea exitosa imprimir: `action: test_echo_server | result: success`, de lo contrario imprimir:`action: test_echo_server | result: fail`.

El script deberá ubicarse en la raíz del proyecto. Netcat no debe ser instalado en la máquina _host_ y no se pueden exponer puertos del servidor para realizar la comunicación (hint: `docker network`). `

### Ejercicio N°3 explicación de la solución:
El script funciona de la siguiente manera:
1- Crea un container temporal con netcat instalado.
2- Duerme un par de segundos para esperar la instalación.
3- Envía un mensaje al servidor a través de la red y luego captura su respuesta en una variable usando docker exec.
4- Elimina el container.
5- Compara la respuesta con el texto enviado. Si son iguales, imprime por consola que la prueba fue un éxito. Caso contrario, imprime que fue un fracaso.

### Dificultades encontradas
-Al principio intenté hacer la prueba a través de los clientes, pero terminé decidiendo que crear un container nuevo era más sencillo y práctico.
-Fue difícil comprender la lógica de conectarme a la red y obtener la ip del servidor. Tuve que leer documentación de docker network.
-Tardé un poco en entender que tenía que conectarme a través del puerto especificado en la configuración del servidor.
-Tuve que aprender a usar variables y ifs en bash, que nunca me había hecho falta hasta ahora.