public function ParseFromFile($file,$date)
    {
        if (!isHoliday($date)) {
            $this->addMemmoryUsage();
            $adjustment = new Query_constructor();
            $symbol_closed_price = new Query_constructor();
            $temp_user_data = new Query_constructor();
            $reorg_user_data = new Query_constructor();

            $command = $this->db->query("DROP TABLE IF EXISTS `users_data_t`");

            $query = "CREATE TABLE IF NOT EXISTS `users_data_t` (
                        `Id` bigint( 20  )  NOT  NULL  AUTO_INCREMENT ,
                        `TradeId` bigint( 20  )  DEFAULT NULL ,
                        `Status` enum(  'normal',  'deleted',  'inserted'  ) DEFAULT  'normal',
                        `ExecDate` date NOT  NULL ,
                        `ExecTime` time NOT  NULL ,
                        `Account` varchar( 50  )  NOT  NULL ,
                        `Side` varchar( 10  )  NOT  NULL ,
                        `Symbol` varchar( 10  )  NOT  NULL ,
                        `Price` varchar( 10  )  NOT  NULL ,
                        `Quantity` int( 20  )  NOT  NULL ,
                        `ContraMMID` varchar( 10  )  NOT  NULL ,
                        `Venue` varchar( 10  )  NOT  NULL ,
                        `InternalRefNumber` int( 10  )  NOT  NULL ,
                        `Liquidity_Flag` varchar( 10  )  NOT  NULL ,
                        `ExecID` int( 20  )  NOT  NULL ,
                        `PerShareChg` varchar( 20  )  NOT  NULL ,
                        `VenueLiqChg` varchar( 20  )  NOT  NULL ,
                        PRIMARY  KEY (  `Id`  ) ,
                        UNIQUE  KEY  `Constraint` (  `ExecDate` ,  `ExecID` ,  `InternalRefNumber`  ) ,
                        KEY  `ExecDate` (  `ExecDate`  ) ,
                        KEY  `Account` (  `Account`  ) ,
                        KEY  `Symbol` (  `Symbol`  ) ,
                        KEY  `TradeId` (  `TradeId`  )  ) ENGINE  = InnoDB  DEFAULT CHARSET  = utf8;";

            $cmd = $this->db->query($query);
            $command = $this->db->query("ALTER TABLE  `users_data_t` CHANGE  `Liquidity_Flag`  `Liquidity_Flag` VARCHAR( 10 ) CHARACTER SET latin1 COLLATE latin1_general_cs NOT NULL ;");

            $allowed_users_q = $this->db->query("SELECT `Account` FROM `bf_allowed_users`")->result_array();

            foreach($allowed_users_q as $allowed_user){
                $allowed_users[$allowed_user['Account']] = $allowed_user['Account'];
            }

            if (!$cmd) {
                return FALSE;
            }
            $start = date("Y-m-d H:i:s");
            foreach ($file as $value) {

                $values = explode(",", trim($value));
                if ($values[0] == 'ExecDate' OR $values[0] == 'PositionID' OR $values[0] == '') {
                    continue;
                }

                if (!isset($allowed_users[str_replace(' ', '', $values[3])]))
                    continue;

                if (count($values) > 9) {
                    if (strcmp(trim($values[20]), ".000000") == 0 || $values[8] == 'LSTK' || $values[8] == 'BSTK') {
                        //TODO REORG

                        $symbol = str_replace(' ', '', $values[5]);
                        $symbol = str_replace('*', '', $values[5]);
                        $parameters_adjustment = array(
                            'Account' => str_replace(' ', '', $values[3]),
                            'ExecDate' => $values[0],
                            'ExecTime' => strotimeformat($values[1]),
                            'Symbol' => $symbol,
                            'Quantity' => $values[7],
                            'ExecID' => $values[12],
                            'Fee' => $values[14],
                        );
                        $adjustment->set($parameters_adjustment);
                        continue;
                    }
                }
                //unrealized
                if (count($values) == 9) {
                    $account = str_replace(' ', '', $values[3]);
                    $symbol = str_replace(' ', '', $values[4]);
                    $position = $values[5];
                    $price = $values[6];
                    if ($position == 'DIS')
                        print_r($values);

                    /*$closed_price_parametrs = array(
                        'Date' => $date,
                        'Account' => $account,
                        'Symbol' => $symbol,
                        'Position' => $position,
                        'Price' => $price,
                    );
                    //$symbol_closed_price->set($closed_price_parametrs);*/
                } else {
                    if ($values[10] == '')
                        $values[10] = rand(1000000000, 2000000000);

                    $symbol = str_replace(' ', '', $values[5]);

                    if (strpos($symbol, "*") > 0) {
                        $this->addHistory("Callable Symbol:" . $symbol);
                    }
                    $symbol = str_replace('*', '', $values[5]);

                    $exec_obj = array(
                        'TradeId' => '0',
                        'Status' => 'normal',
                        'ExecDate' => $values[0],
                        'ExecTime' => strotimeformat($values[1]),
                        'Account' => str_replace(' ', '', $values[3]),
                        'Side' => $values[4],
                        'Symbol' => $symbol,
                        'Price' => round($values[6],4),
                        'Quantity' => $values[7],
                        'ContraMMID' => $values[8],
                        'Venue' => $values[9],
                        'InternalRefNumber' => $values[10],
                        'Liquidity_Flag' => $values[11],
                        'ExecID' => $values[12],
                        'PerShareChg' => $values[15],
                        'VenueLiqChg' => $values[17]);

                    if ($values[0] != $date) {
                        $exec_obj['Date'] = $date;
                        $reorg_user_data->set($exec_obj);
                    } else {
                        $temp_user_data->set($exec_obj);
                    }

//                    echo $values[6].'</br>';

                }
            }

            $this->addTimeLog($start, 'Parse file');
            $this->addMemmoryUsage();
            $reorg_user_data->insert('bf_users_data_reorg');
            unset($reorg_user_data);

            $adjustment->insert('bf_adjustment');
            unset($adjustment);

            //$symbol_closed_price->insert('bf_symbols_closed_price');
            //unset($symbol_closed_price);

            $temp_user_data->insert('users_data_t');
            //Update ECN
            $this->db->query("UPDATE `users_data_t` as d
                INNER JOIN `bf_ecn_fee` as e
                ON d.`ContraMMID` = e.`ContraMMID` and d.`Liquidity_Flag` = e.`Liquidity_Flag`
                SET d.`VenueLiqChg`=round(d.`Quantity`*e.`VenueLiqChg`,4)
                WHERE d.`ExecDate` = '" . $date . "'");

            unset($temp_user_data);
            $this->addMemmoryUsage();

            $start = date("Y-m-d H:i:s");
            $dateExecs = $this->db->query("SELECT * FROM `users_data_t` order by ExecDate,ExecTime,InternalRefNumber,ExecID")->result_array();

            $this->addTimeLog($start, 'Get data from temp table');
            $this->addMemmoryUsage();

            if (count($dateExecs) == 0) {
                $this->addHistory("file is empty");
            }
            $start = date("Y-m-d H:i:s");
            $dateClosePrice = $this->db->query("SELECT `Symbol`,`Price` FROM `bf_symbols_closed_price` WHERE `Date`=? GROUP BY `Symbol`", array($date))->result_array();
            $symbol_price = array();
            foreach ($dateClosePrice as $row) {
                $symbol_price[$row['Symbol']] = $row['Price'];
            }
            $this->addTimeLog($start, 'Get closed price');
            $this->addMemmoryUsage();

            $this->ParseFromArrayForDate($date, $dateExecs, $symbol_price);
            $command = $this->db->query("DROP TABLE IF EXISTS `users_data_t`");
            unset($dateExecs);
            unset($symbol_price);

            $time = strtotime($date);
            $date2 = date('Ymd', $time);
            $file_name = 'EOD_RussianGrp_' . $date2 . '.csv';
            setFileAsProcess($date, $file_name);
        }
    }