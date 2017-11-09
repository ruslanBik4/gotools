unit test;

interface

uses
  Classes, SysUtils, janXMLParser2, BaseClassUnit, DataPacketUnit,
  NetConfigClassUnit, RNetConfigClassUnit, DataBaseClassUnit, RDataBaseClassUnit,
  QueueExchTaskUnit;

const
  ///  ����� ������ ��� ������ TRQueueExchTask
  R_DB_QUEUE_EXCH_TASK_STREAM = 101;

type

  { TRQueueExchTask }

  ///  ����� ������������ ���:
  ///  1) ������ � ��������� ����������� ����������� � ������� ������� ��������
  ///  2) ��� �������� ������ � ���� BLOB (������� ������) � ������ push
  ///  �� ��������� ���������� ������� ����� ����� �� ����� � ���������� �����
  ///  ����������� (�������� TDataPacket) �������������� ����� �������� ����� �����
  ///  TRQueueExchTask <--->  TRQueueExchTask
  ///
  ///  ���������������:
  ///  ����� ������������ ���������� � ����������� ���������� ���� FContentCashe
  ///  �, ���� ������� � ��� ����� ������� ��� ��������� ������� (��� ���� ��������� ������� �������� ��  FNetConfig!),
  ///  �� ��������� ����� TDataPacket � ����������� � ����� �������� � �������� ��� � ���� ������� ��������
  ///  ����������� ����� �� ��� ������� ������ �����, �������� ���� ����� � ����������
  ///  ���������� � ������� �������� � ���� ��������� ���.
  ///  ���� ����� ������ ���������� �������� � ���� �������������� �������, ������� ������
  ///  ���� ������� �� ��������� ��������� � ������ push (��� ���� ������������ � FNetConfig),
  ///  �� ����� ���������� � ���� �� ������� ������ ��������, ��������� �� ���� ����� TDataPacket
  ///  � ���������� ����� � ���� ������� �������� �� ��������� ���������
  ///
  TRQueueExchTask = class(TCashedQueueExchTask)
  private
    /// �������� ����� ���� �� ����� GER/LER, ����� ������� �������� ��� ������ (���� <link> ����� GER/LER)
    FNetLink : TRNetLink;
    // �-��� ��������� �������� �� �� ������� ��������� ������
    //function IsSubscribed(Content : TRContent) : Boolean;
    // �-��� ��������� ������ �� ������� ���� ������� � ������ push
    //function IsPush(Content : TRContent) : Boolean;
  protected
    //function PrepareFilter : string; override;
    ///  ������� ��� ��������� ��������� ������
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

  // ������� ������� NetLink, ������� ������� MyHostName � RemoteHostName
  // ��������� NetLink ������������ ��� ���������� ��������
  // � ��� ����������� �����, � ������� �������� ������� �����
  function GetMyNetLink() : TRNetLink;
  var
    List : TList;

  begin
    List := TList.Create;
    try
      /// �������� ������ ���� ������, ������� �������� ��� ��������� � ��������� ���������
      FNetConfig.GetServersNetLink(MyHostName, RemoteHostName, List);

      if List.Count > 0 then
        Result := TRNetLink(List[0])   ///  ??? ������ ������ ���� � ������ - ���? ������� ��� ������!
      else
        Result := nil;
    finally
      List.Free;
    end;
  end;

begin
  inherited;
  ///  ������������� ����� ������ �����
  StreamNo := R_DB_QUEUE_EXCH_TASK_STREAM;
  ///  ���� ���� ����� ����� � ������ ���� ������ ����� ���������� � ���������� ����������
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
  // ������� ��������������� ������� �����
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
  // ���� �� ��� ��� �� ����� ��������, ��
  // ������ �������� ������������ ������
  if Result = '' then
    Result := '1 = 0';
  Result := '(CONT.jid is not null) and (' + Result + ')';
end;
}

///  ������� ��� ��������� ��������� ������
///  ������� ���������� ������� ��� ��������� ������ ������ TDataPacket �� ���������� ����������
///  ���� ADataPacket ������������ ��������� - �� ���� ���������� true
function TRQueueExchTask.ProcessingDataPacket(ADataPacket : TDataPacket) : Boolean;
const
  SRecSaved = '� ��� ��������� %s ����� �������';

var
  //Content : TRContent;
  Data: string;
  List: TContentList;
  i : Integer;

begin
///  � ���������� TDataPacket ����� ���������:
///  1) ������ ������ �������� � ���������� ������� ���
///  2) �������, �� ������� �� ��������� � ������ push
///
///  ������� ��������� ��� ������ ADataPacket � ��������� �� ��� TRContent ��� ������
///  � �������������� ������ PutDataExt � ��������� ����� ������ �����.

  Result := True;   ///  ����� ����� ���������
  try
    /// �������� ������ ������ � ���� ������
    Data := ADataPacket.DataStr;
    ///  ������� ������ �������� � �������� ���� TRContent �� ���������� ������
    List := TContentList.Create(TRContent, Data);

    ///  ���� � ������� ���� ������� - �� ���������� ��� � ���
    if List.Count > 0 then
    begin
      for i := 0 to List.Count - 1 do
        ///  ������ ��������� �������� ���������� �������� � ����������� � �����
        TRDataBase(Cashe).PutDataExt(FNetLink.NetLinkId, List[i]);

      ///  �������� ���������� ���������� (��� ������������)
      DoTaskEvent(Format(SRecSaved, [IntToStr(List.Count)]));
      ///  ���������� � ��� ����������� � ������ � ���
      WriteLog(ltDebug, Format(SRecSaved, [IntToStr(List.Count)]));
    end;
  except
    Result := False;
  end;
end;

function TRQueueExchTask.Tick:Boolean;
const
  SRecFound = '� ���� ������� %s ����� �������';

var
  ContentList : TContentList;
  Content : TRContent;
  Packet: TDataPacket;
  i, AvailCount: Integer;
  strJSON: string;

begin
///  ������������ ���������� � ���� � �������� ������ ���������� � �����
///  ����������� � ������� ���������� ��������� ��������
///  �������� �������������� ������� � ���� TContectList
///  �� ����������� ������ �������� ������� ������� ������ ���� ������� � ������
///  push � ���������� � ���� ��� ��������� ������ ����� ��������
///  (� ������ push �������� �������� ������ �������� ��������������� � ���� �����)
///  �� ���������� � ����� ������ � ������ �� �������� � ������ push �����������
///  ���� �����  TDataPacket � ������ ����������� � ������������ � �������


  Result := true;

  ///  ���� ��� �� ���������� - ������ �� ������
  if not Assigned(Cashe) then
    Exit;

  ///  ��������� ������� ������� ��� ������ � �������
  AvailCount := MaxQueueSize - FPacketQueue.Count;
  ////  ���� ��� ���� ����� � ������� ->
  if (AvailCount > 0) then
  try
    // ��������� ���
    ContentList := TContentList.Create;
    try
      ///  �������� � ContentList ���������� ������� � ������� LastRowId � ���������� �� ������ AvailCount
      ///  ����� ������� ��� ����� ������������ Filter
      Cashe.GetDataList(Filter, LastRowId, AvailCount, ContentList);

      i := 0;
      while i < ContentList.Count do
      begin
        Content := TRContent(ContentList[i]);
        ///  ������������� ��������� ������� ������
        if LastRowId < Content.RowId then
          LastRowId := Content.RowId;

        // ��������� �������� �� �� ���� ������� ��������� ������
        if FNetLink.IsSubscribed(RemoteHostName, Content.Key) then
        begin
          // ������ ���������, ��� ������� ������ ���� ������� � ������ push,
          if FNetLink.IsPush(RemoteHostName, Content.Key) then
            ///  ���� ��������� ������ �������� �� ���� ������� � ������ push �� ����������� ������ ��������
            Cashe.GetData(Content.Hash, TContent(Content));

          Inc(i);
        end
        else
        begin
          // ���� �� �������� �� �� ���� ������� ��������� ������ �� ������� ������� �� ������
          ContentList.Remove(Content);
          Content.Free;
        end;
      end;

      ///  � ����� � ���� ����� � ContentList ��������� ������ �������, �� ������� �������� ��������� ���������
      ///  � ���� ��� �������� � ������ push - �� ��� � ����� � �������
      ///  ���� ���� ������� �� �������� �� ��������� ���������
      if ContentList.Count > 0 then
      begin
        ///  ���������� �� ���� (��� �������)
        DoTaskEvent(Format(SRecFound, [IntToStr(ContentList.Count)]));
        ///  ���������� �� ���� (��� �������)
        WriteLog(ltDebug, Format(SRecFound, [IntToStr(ContentList.Count)]));
        // ��������� ���� ����� �������� ������ TDataPacket �� ��� ��������
        Packet := TDataPacket.Create(FStreamNo);
        ///  ������ ������ - �������� ������
        Packet.Mission := MIS_DATA;
        ///  ��������� ������� � ���� ������ JSON
        strJSON := ContentList.ToJSON();
        //Packet.PutData(Data);
        Packet.DataStr := strJSON;
        ///  ���������� ����� � �������
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
  ///  ���� ����� �� ����� ������������ �������
  Result := False;
end;

end.
