rm -rf ~/.dobby

if grep -q "source ~/.dobby/scripts/register.sh" ~/.zshrc; then
  sed -i '' '/source ~\/.dobby\/scripts\/register.sh/d' ~/.zshrc
fi
