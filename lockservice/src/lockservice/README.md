Serwer rygli
============

Zadanie to polega na napisaniu serwera udostępniającego zdalnym klientom mutexy.
Serwer posiada kolekcję mutexów o nazwach będących ciągami znaków i umożliwia
klientom ich zamykanie i odmykanie.

Zakładamy, że serwer nie ulega awariom; jednak sieć między klientami a serwerem jest zawodna. Każde żądanie wysłane przez klienta może
się powieść, ale też może nigdy nie dotrzeć do serwera (i na kliencie zakończyć się błędem), jak również dotrzeć do serwera, być
przez niego wykonane, ale zakończyć się błędem na kliencie (odpowiada to sytuacji, gdy odpowiedź serwera została zgubiona z powodu
zawodności sieci). Można założyć, że w każdym momencie dla każdego klienta jest tak, że jeśli będzie próbował wykonać jakieś wywołania RPC,
to w końcu któreś się uda.

W plikach src/lockservice/{common,server,client}.go znajduje się implementacja serwera i klienta, pozbawiona implementacji Unlock. Implementacja ta działa w obliczu
braku awarii. Waszym zadaniem jest tak ją poprawić, by obsługiwała również wywołanie Unlock oraz by działała też w obliczu awarii.


Interfejs
---------

Waszym zadaniem jest uzupełnienie funkcjonalności zaimplementowanej w plikach client.go, server.go i common.go.
Interfejs, który będzie testowany to:

* func NewLockService() -- metoda, która winna zwrócić serwer w postaci obiektu, który można zarejestrować w rpc.Server
* func NewClient(ls) -- metoda, która winna zwrócić obiekt klienta (patrz niżej) używający podanego jako argument połączenia
  z serwerem. Argument ten jest obiektem z interfejsem identycznym jak interfejs rpc.Client
* type Client -- obiekt reprezentujący klienta. Powinien on implementować poniższe metody. Można założyć, że nie
  będą one wywoływanie jednocześnie.

  * func Lock(name) -- zamknij rygiel o nazwie name, jeśli jest otwarty. Zwróć true jeśli był otwarty, false wpp.
  * func Unlock(name) -- otwórz rygiel o nazwie name. Można założyć, że rygiel ten jest zamknięty w momencie wywołania
    tej metody i nikt nie próbuje go otworzyć jednocześnie.
  * func Close() -- zakończ działanie klienta. Metoda ta powinna wywołać metodę Close() na obiekcie klienta RPC, którego
    ten klient używa, oraz dokonać wszelkiej niezbędnej finalizacji.

Uwaga
-----

*  Serwer RPC jest inherentnie wielowątkowy: każde żądanie wykonuje się niezależnie, być może współbieżnie. Trzeba to brać pod uwagę podczas manipulacji stanem serwera.
*  Wywołanie Lock i sparowane z nim wywołanie Unlock mogą pochodzić od różnych klientów.
*  Możecie założyć, że stan generatora liczb losowych nigdy nie będzie taki sam w procesach różnych klientów.

Technikalia
-----------

Aby przetestować swoje rozwiązanie, należy uruchomić załączone unit testy. Czyni się to poleceniem

go test -race

wykonanym w katalogu src/lockservice za pomoca shella z poprawnie ustawionym GOPATH. Aby ustawić GOPATH, należy wykonać

source ./activate

Na satori należy wysłać spakowany katalog src/lockservice (pliki z testami można pominąć: będą ignorowane).
