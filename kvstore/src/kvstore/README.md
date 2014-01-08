Master-slave: pierwsze podejście
================================

Zadanie to polega na napisaniu serwisu udostępniającego zdalnym klientom słownik. Serwer umożliwia klientom czytania wartości ze słownika
i ustawianie wartości w słowniku. Serwis ten ma się składać z jednego serwera mastera i dowolnej liczby serwerów slave. Wszystkie z nich
są cały czas sprawne (na razie). Tak jak w poprzednim zadaniu, sieć między nimi nie jest pewna i może mieć przejściowe awarie.

Aby przygotować się do kolejnego zadania, w którym awarie serwerów będą już mogły następować, serwis ten musi spełniać następujące warunki:
* Każdy slave musi trzymać cały stan słownika. Musi on trzymać ,,prawdziwą'' wersję owego, czyli uzyskaną przez zaaplikowanie jakiegoś prefiksu wszystkich rozpoczętych operacji.
* Zanim master udzieli odpowiedzi klientowi, każdy slave musi mieć w swoim słowniku wersję, która uwzględnia operację zleconą przez klienta.

W pliku src/kvstore/server.go znajduje się szkielet implementacji serwera. Tym razem nie musicie pisać klienta, wasz serwer ma jedynie implementować następujące
wywołania RPC:
* Get(GetRequest, GetResponse) -- czyta wartość spod klucza Key i zwraca ją, lub nil gdy takowa nie istnieje
* Set(SetRequest, SetResponse) -- zapisuje wartość Value pod klucz Key

Master winien implementować obie te metody. Slave winien implementować metodę Get i zwracać swój bieżący stan wiedzy w odpowiedzi na nią.

Interfejs
---------

Interfejs, który będzie testowany to:

* func NewSlave(quit) -- metoda, która winna zwrócić serwer slave, gotowy do zarejestrowania w rpc.Server
* func NewMaster(slaves, quit) -- metoda, która winna zwrócić serwer master, gotowy do zarejestrowania w rpc.Server. Pierwszy argument to lista serwerów slave należących do serwisu.

Obie te metody otrzymują też jako argument kanał. Kanał ten zostanie zamknięty gdy serwer powinnien zakończyć działanie. Jego obsługa nie jest wymagana,
ale jej brak utrudnia debugowania (poprzez potencjalne pozostawianie żywych, zbędnych goroutines) oraz może powodować interferencję między różnymi testami.

Uwaga
-----

*  Serwer RPC jest inherentnie wielowątkowy: każde żądanie wykonuje się niezależnie, być może współbieżnie. Trzeba to brać pod uwagę podczas manipulacji stanem serwera.
*  Sugeruję zaimplementować serwery master i slave za pomocą tego samego obiektu; ułatwi to kolejne zadanie, w którym ich role będą mogły się zmieniać.

Technikalia
-----------

Aby przetestować swoje rozwiązanie, należy uruchomić załączone unit testy. Czyni się to poleceniem

go test -race -cpu 4

wykonanym w katalogu src/lockservice za pomoca shella z poprawnie ustawionym GOPATH. Aby ustawić GOPATH, należy wykonać

source ./activate

Na satori należy wysłać spakowany katalog src/kvstore (pliki z testami można pominąć: będą ignorowane).
