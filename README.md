# About PR-StatusUpdater

Telegram bot for automating the collection and evaluation of Pull Requests from any git platform (GitHub, GitLab..)

**Just two clicks:**

<div align="center">
  <img width="535" height="491" alt="image" src="https://github.com/user-attachments/assets/a87a1392-5b0f-440a-b22b-096c9c6be081" />
</div>

**And you get csv with link:**

<div align="center">
  <img width="546" height="436" alt="image" src="https://github.com/user-attachments/assets/afb5b470-75cd-463d-a3f0-0619b0849093" />
</div>


### What is it?

This utility, in the form of a Telegram bot, can fetch pull requests from any git platform into a plain JSON, apply certain policies to them, and upload them as a gist/snippet for further export, for example, to Google Sheets as CSV (you get access to the raw data link).

### Who might find this utility useful?

It is primarily suitable for teachers who introduce students to git and have them submit lab assignments and exams via GitHub, for example. If you don't want to manually track deadlines, grades, compile reports, and so on, this utility is for you.

### What policies can be applied to pull requests?

- **You can set deadlines for pull request creation and submission**  
    In fact, you can specify any timeframes. For instance, a teacher can set a `fine` label on GitHub when they consider a lab assignment ready for submission.
- **You can use labels to penalize or reward students**  
    Using labels like `+1` or `-2`, etc., you can influence the final **score** students will have.

### What is NOT available at the moment?

The program is not finished, but it already allows fetching pull requests via GitHub. 
Currently, **it is not possible to**:

- Use GitLab or Bitbucket (but you can implement it yourself! The program provides an interface and a great factory for that).
    
- Configure the application via the bot. For now, you need to edit the configuration files directly and rewrite the YAML files. However, all comments are included.
    
- Fetch more than 100 pull requests per request (GitHub). Unfortunately, this is a GraphQL limitation, and I'm still thinking about a solution.
    
- Run the program NOT via Telegram. I have abandoned the console version, maybe in future...
    
# Installation

1. Clone the master branch of this repository.
2. Create a `.env` file in the project root:
```
TELEGRAM_BOT_TOKEN=<your_bot_token>
SETTINGS_WHITELIST=<your_id>,<colleague_id>
```
- To fill in `TELEGRAM_BOT_TOKEN`, you need to create your own Telegram bot, for example, via [https://t.me/BotFather](https://t.me/BotFather). It will provide you with a token to insert into this field.
- `SETTINGS_WHITELIST` is for administrators; only they have the right to configure the bot via Telegram.
3. If you are using the utility with GitHub, install `gh cli`. IMPORTANT: authenticate in it! **Generate an access key that allows it to modify gists!**
4. Create repositories from which you will fetch pull requests. Add them to the program via the Telegram bot or the `repositories.yaml` file. Write your YAML configuration for GitHub requests. Examples are available in the code, both simple and slightly more complex.
5. Inside, there is an executable file `pr_updater`, built on Ubuntu 20.04.6 LTS. You can run it. If it doesn't work, **it is not hard to build it yourself**:
    - Install Go.
    - Run `go build -o pr_updater cmd/service/main.go`.
    - Run `pr_updater`.
    - Done ;-)
