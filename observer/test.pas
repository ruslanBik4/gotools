unit test;

interface

uses
  Classes, SysUtils, janXMLParser2, BaseClassUnit, DataPacketUnit,
  NetConfigClassUnit, RNetConfigClassUnit, DataBaseClassUnit, RDataBaseClassUnit,
  QueueExchTaskUnit;

const
  ///  номер потока для задачи TRQueueExchTask
  R_DB_QUEUE_EXCH_TASK_STREAM = 101;

type

  { TRQueueExchTask }

  ///  класс предназначен для:
  ///  1) обмена с удаленным комсервером информацией о наличии нужного контента
  ///  2) для передачи данных в виде BLOB (массива данных) в режиме push
  ///  На удаленном комсервере запущен ТОЧНО ТАКОЙ ЖЕ КЛАСС и фактически обмен
  ///  информацией (пакетами TDataPacket) осуществляется двумя классами между собой
  ///  TRQueueExchTask <--->  TRQueueExchTask
  ///
  ///  Функионирование:
  ///  Класс периодически обращается к глобальному локальному кешу FContentCashe
  ///  и, если находит в нем новый контент для удаленной стороны (что надо удаленной стороне известно из  FNetConfig!),
  ///  то формирует пакет TDataPacket с метаданными о новом контенте и помещает его в свою очередь передачи
  ///  Аналогичный класс на той стороне канала связи, получает этот пакет и записывает
  ///  метаданные о наличии контента в свой локальный кеш.
  ///  Если среди нового локального контента в кеше обнаруживается контент, который должен
  ///  быть передан на удаленный комсервер в режиме push (это тоже определяется в FNetConfig),
  ///  то класс обращается к кешу за данными такого контента, формирует из него пакет TDataPacket
  ///  и записывает пакет в свою очередь передачи на удаленный комсервер
  ///
  TRQueueExchTask = class(TCashedQueueExchTask)
  private
    /// описание линка сети из файла GER/LER, через который работает эта задача (нода <link> файла GER/LER)
    FNetLink : TRNetLink;
    // ф-ция проверяет подписан ли на контент удаленный сервер
    //function IsSubscribed(Content : TRContent) : Boolean;
    // ф-ция проверяет должен ли контент быть передан в режиме push
    //function IsPush(Content : TRContent) : Boolean;
  protected
    //function PrepareFilter : string; override;
    ///  функция для обработки входящего пакета
    function ProcessingDataPacket(ADataPacket : TDataPacket) : Boolean; override;

  public
    constructor Create(const AMyName, AMyHostName, ARemoteHostName: string;
      ANetConfig: TNetConfig); override;
    function Tick : Boolean; override;
    function DoCmd(ACmd : TjanXMLNode2 ): Boolean; override;
  end;

implementation

{ TRQueueExchTask }

constructor TRQueueExchTask.Create(const AMyName, AMyHostName, ARemoteHostName: string;
  ANetConfig: TNetConfig);

  // функция находит NetLink, которым связаны MyHostName и RemoteHostName
  // найденный NetLink используется для фильтрации контента
  // и для определения линка, с помощью которого получен пакет
  function GetMyNetLink() : TRNetLink;
  var
    List : TList;

  begin
    List := TList.Create;
    try
      /// получаем список всех линков, которые связыают наш комсервер и удаленный комсервер
      FNetConfig.GetServersNetLink(MyHostName, RemoteHostName, List);

      if List.Count > 0 then
        Result := TRNetLink(List[0])   ///  ??? почему первый линк в списке - наш? кажется тут ошибка!
      else
        Result := nil;
    finally
      List.Free;
    end;
  end;

begin
  inherited;
  ///  устанавливаем номер потока таска
  StreamNo := R_DB_QUEUE_EXCH_TASK_STREAM;
  ///  ищем линк этого таска в списке всех линков этого комсервера и удаленного комсервера
  FNetLink := GetMyNetLink();
end;

{
function TRQueueExchTask.IsSubscribed(Content : TRContent) : Boolean;
var
  i : Integer;
begin
  for i := 0 to FNetLink.SubscribeCount - 1 do
    if (FNetLink.Subscribes[i].TargetId = RemoteHostName) and
       (FNetLink.Subscribes[i].IsMatch(Content.Key)) then
    begin
      Result := True;
      Exit;
    end;
  Result := False;
end;

function TRQueueExchTask.IsPush(Content : TRContent) : Boolean;
var
  i : Integer;
begin
  for i := 0 to FNetLink.SubscribeCount - 1 do
    if (FNetLink.Subscribes[i].TargetId = RemoteHostName) and
       (FNetLink.Subscribes[i].IsMatch(Content.Key)) and
       (FNetLink.Subscribes[i].Push) then
    begin
      Result := True;
      Exit;
    end;
  Result := False;
end;

function TRQueueExchTask.PrepareFilter : string;
var
  i : Integer;
begin
  Result := '';
  // находим противоположную сторону линка
  if Assigned(FNetLink) then
  begin
    for i := 0 to FNetLink.SubscribeCount - 1 do
      if (FNetLink.Subscribes[i].TargetId = RemoteHostName) and
         (FNetLink.Subscribes[i].Filter <> '') then
      begin
        if Result <> '' then
          Result := Result + ' OR ';
        Result := Result + '(key = ' + QuotedStr(FNetLink.Subscribes[i].Filter) + ')';
      end;
  end;
  // если на нас нет ни одной подписки, то
  // задаем заведомо невыполнимый фильтр
  if Result = '' then
    Result := '1 = 0';
  Result := '(CONT.jid is not null) and (' + Result + ')';
end;
}

///  функция для обработки входящего пакета
///  функция вызывается предком при получении нового пакета TDataPacket от удаленного комсервера
///  если ADataPacket обрабатывает нормально - то надо возвращать true
function TRQueueExchTask.ProcessingDataPacket(ADataPacket : TDataPacket) : Boolean;
const
  SRecSaved = 'В кеш добавлено %s новых записей';

var
  //Content : TRContent;
  Data: string;
  List: TContentList;
  i : Integer;

begin
///  В полученных TDataPacket может находится:
///  1) список нового контента с удаленного сервера или
///  2) контент, на который мы подписаны в режиме push
///
///  Фцнкция разбирает все записи ADataPacket и формирует из них TRContent для записи
///  с использованием метода PutDataExt с указанием имени нашего линка.

  Result := True;   ///  пакет будет обработан
  try
    /// получаем данные пакета в виде строки
    Data := ADataPacket.DataStr;
    ///  создаем список контента с записями типа TRContent из полученной строки
    List := TContentList.Create(TRContent, Data);

    ///  если в списоке есть контент - то записываем его в кеш
    if List.Count > 0 then
    begin
      for i := 0 to List.Count - 1 do
        ///  каждый экземпляр контента записываем отдельно с информацией о линке
        TRDataBase(Cashe).PutDataExt(FNetLink.NetLinkId, List[i]);

      ///  передаем отладочную информацию (для тестирования)
      DoTaskEvent(Format(SRecSaved, [IntToStr(List.Count)]));
      ///  записываем в лог инфомраицию о записи в кеш
      WriteLog(ltDebug, Format(SRecSaved, [IntToStr(List.Count)]));
    end;
  except
    Result := False;
  end;
end;

function TRQueueExchTask.Tick:Boolean;
const
  SRecFound = 'В кеше найдено %s новых записей';

var
  ContentList : TContentList;
  Content : TRContent;
  Packet: TDataPacket;
  i, AvailCount: Integer;
  strJSON: string;

begin
///  Периодически обращается к кешу и получает список информации о новом
///  поступившем с момента последнего обращения контенте
///  согласно установленного фильтра в виде TContectList
///  Из полученного списка выбирает контент который должен быть передан в режиме
///  push и обращается к кешу для получения данных этого контента
///  (В режиме push возможна передача только контента представленного в виде блоба)
///  Из информации о новых данных и данных на передачу в режиме push формируется
///  один пакет  TDataPacket с высшим приоритетом и записывается в очередь


  Result := true;

  ///  если кеш не установлен - ничего не делаем
  if not Assigned(Cashe) then
    Exit;

  ///  вычисляем сколько пакетов еще влезет в очередь
  AvailCount := MaxQueueSize - FPacketQueue.Count;
  ////  если еще есть место в очереди ->
  if (AvailCount > 0) then
  try
    // мониторим кэш
    ContentList := TContentList.Create;
    try
      ///  получаем в ContentList метаданные контета с позиции LastRowId в количестве не больше AvailCount
      ///  какой контент нам нужен определяется Filter
      Cashe.GetDataList(Filter, LastRowId, AvailCount, ContentList);

      i := 0;
      while i < ContentList.Count do
      begin
        Content := TRContent(ContentList[i]);
        ///  устанавливаем последнюю позицию чтения
        if LastRowId < Content.RowId then
          LastRowId := Content.RowId;

        // проверяем подписан ли на этот контент удаленный сервер
        if FNetLink.IsSubscribed(RemoteHostName, Content.Key) then
        begin
          // теперь проверяем, что контент должен быть передан в режиме push,
          if FNetLink.IsPush(RemoteHostName, Content.Key) then
            ///  если удаленный сервер подписан на этот контент в режиме push то запрашиваем данные конетнта
            Cashe.GetData(Content.Hash, TContent(Content));

          Inc(i);
        end
        else
        begin
          // если НЕ подписан ли на этот контент удаленный сервер то удаляем контент из списка
          ContentList.Remove(Content);
          Content.Free;
        end;
      end;

      ///  в итоге в этой точке в ContentList останется только контент, на который подписан удаленный комсервер
      ///  а если еще подписан в режиме push - то еще и сразу с данными
      ///  если есть контент на передачу на удаленный комсервер
      if ContentList.Count > 0 then
      begin
        ///  уведомляем об этом (для отладки)
        DoTaskEvent(Format(SRecFound, [IntToStr(ContentList.Count)]));
        ///  уведомляем об этом (для отладки)
        WriteLog(ltDebug, Format(SRecFound, [IntToStr(ContentList.Count)]));
        // формируем один пакет передачи данных TDataPacket на все элементы
        Packet := TDataPacket.Create(FStreamNo);
        ///  миссия пакета - передача данных
        Packet.Mission := MIS_DATA;
        ///  считываем контент в виде строки JSON
        strJSON := ContentList.ToJSON();
        //Packet.PutData(Data);
        Packet.DataStr := strJSON;
        ///  записываем пакет в очередь
        FPacketQueue.PutPacket(Packet);
        Packet.Free;
      end;
    finally
      ContentList.Free;
    end;

  except
    Result := False;
  end;

end;

function TRQueueExchTask.DoCmd(ACmd : TjanXMLNode2 ): Boolean;
begin
  ///  этот класс не умеет обрабатывать команды
  Result := False;
end;

end.
