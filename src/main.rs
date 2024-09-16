use std::hash::Hash;
use teloxide::{prelude::*, utils::command::BotCommands};
use solana_sdk::signer::keypair::Keypair;
use solana_sdk::signer::Signer;
use std::sync::Arc;
use rusqlite::{params, Connection, Result};

#[derive(BotCommands, Clone)]
#[command(rename_rule = "lowercase", description = "These commands are supported:")]
enum Command {
    #[command(description = "start this game ...")]
    Start,
    #[command(description = "display this text.")]
    Help,
    #[command(description = "handle a username.")]
    Username(String),
    #[command(description = "handle a username and an age.", parse_with = "split")]
    UsernameAndAge { username: String, age: u8 },
}

async fn answer(bot: Bot, msg: Message, cmd: Command) -> ResponseResult<()> {
    let db = Arc::new(Connection::open("data.db").unwrap());
    db.execute(
        "CREATE TABLE IF NOT EXISTS keys (,
            user_id INTEGER PRIMARY KEY,
            user_addr TEXT NOT NULL,
            user_key TEXT NOT NULL,
        )",
        [],
    ).unwrap();
    match cmd {
        Command::Start => {
            let user_id_to_check = msg.from.unwrap().id.0;

            // Check if the user_id exists
            match db.prepare("SELECT EXISTS(SELECT 1 FROM user_data WHERE user_id = ?)") {
                Ok(mut stmt) => {
                    match stmt.query_row(params![user_id_to_check], |row| row.get::<usize, bool>(0)) {
                        Ok(_user_id) => {
                            bot.send_message(msg.chat.id, "User ID exists in the table.").await?;
                        }
                        Err(_) => { println!("Error getting user id"); },
                    }
                }
                Err(_) => {println!("err")}
            }
            // let kp = Keypair::new();
            // let mnemonic = mnemonic::to_string(kp.to_bytes());
            // bot.send_message(msg.chat.id, kp.pubkey().to_string()).await?;
            // bot.send_message(msg.chat.id, kp.to_base58_string()).await?;
            // bot.send_message(msg.chat.id, mnemonic).await?
            bot.send_message(msg.chat.id, "User ID not exists in the table.").await?
        },
        Command::Help => bot.send_message(msg.chat.id, Command::descriptions().to_string()).await?,
        Command::Username(username) => {
            bot.send_message(msg.chat.id, format!("Your username is @{username}.")).await?
        }
        Command::UsernameAndAge { username, age } => {
            bot.send_message(msg.chat.id, format!("Your username is @{username} and age is {age}."))
                .await?
        }
        _ => {
            bot.send_message(msg.chat.id, "i dont know your command ...".to_string())
                .await?
        }
    };

    Ok(())
}

#[tokio::main]
async fn main() {
    pretty_env_logger::init();
    log::info!("Starting command bot...");

    let bot = Bot::from_env();

    Command::repl(bot, answer).await;
}